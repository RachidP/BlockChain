package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST" //verify if the blockchain exist.
	genesisData = "First transaction from Genesis"
)

// BlockChain rappresent a BlockChain
type BlockChain struct {
	LastHash []byte //last Hash of the last block in the chain
	Database *badger.DB
}

//BlockChainIterator iterate over a blockchain
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// InitBlockChain build our initial BlockChian.
func InitBlockChain(address string) *BlockChain {
	var lastHash []byte
	if DBExist() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	//set path where to save data
	opts := badger.DefaultOptions
	opts.Dir = dbPath      // where the db store the keys and metadata
	opts.ValueDir = dbPath //where the db will store all the values

	//open the db
	db, err := badger.Open(opts)
	HandleErr(err)

	//db.Update access to db for reading a writing transaction

	err = db.Update(func(txn *badger.Txn) error {
		//lh := is the key (last hash)

		//no existing blockchain in our db
		fmt.Println("No existing blockchain found in the db")
		cbtx := CoinbaseTx(address, genesisData)
		genesis := Genesis(cbtx) //make genesis block
		fmt.Println("Genesis Created")

		//uses the genesis hash as the key of the genesis hash, serialize the block and put into db
		err := txn.Set(genesis.Hash, genesis.Serialize())
		HandleErr(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err

	})

	HandleErr(err)
	blockchain := BlockChain{LastHash: lastHash, Database: db}
	return &blockchain
}

// AddBlock add a Block to the BlockChain
func (chain *BlockChain) AddBlock(transactions []*Transaction) *Block {

	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		HandleErr(err)
		lastHash, err = item.Value()
		return err
	})
	HandleErr(err)

	newBlock := CreateBlock(transactions, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {

		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		HandleErr(err)

		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err

	})

	HandleErr(err)
	return newBlock

}

//Iterator convert a BlockChian struct into a BlochainIterator struct
func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := BlockChainIterator{
		CurrentHash: chain.LastHash,
		Database:    chain.Database,
	}
	return &iter

}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		HandleErr(err)
		encodeBlock, err := item.Value()
		block = Deserialize(encodeBlock)
		return err
	})
	HandleErr(err)
	iter.CurrentHash = block.PrevHash
	return block
}

//DbExist check if the DB exist
func DBExist() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true

}
func ContinueBlockChain(address string) *BlockChain {
	if DBExist() == false {

		fmt.Println("No existing Blockchain found, create one!")
		runtime.Goexit()

	}
	var lastHash []byte
	//set path where to save data
	opts := badger.DefaultOptions
	opts.Dir = dbPath      // where the db store the keys and metadata
	opts.ValueDir = dbPath //where the db will store all the values
	//open the db
	db, err := badger.Open(opts)
	HandleErr(err)

	err = db.Update(func(txn *badger.Txn) error {

		// if the blockchain has  already been stored into db
		item, err := txn.Get([]byte("lh"))
		HandleErr(err)
		lastHash, err = item.Value()
		return err

	})

	HandleErr(err)
	chain := BlockChain{LastHash: lastHash, Database: db}
	return &chain

}

//FindUTXO go through all of the transactions and find all the unspent outputs in those transactions.
func (chain *BlockChain) FindUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	//iterate over the blocks
	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXO
}

//FindTransaction get an ID find the trans	action
func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

//SignTransaction sign a transaction
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		HandleErr(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

//VerifyTransaction verify a transaction
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		HandleErr(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
