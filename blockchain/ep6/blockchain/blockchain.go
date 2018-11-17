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
func (chain *BlockChain) AddBlock(transactions []*Transaction) {

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

//FindUnspentTransactions
//Unspent transactions are transactions that have output wich are not referenced by other inputs
//that's means that are tokens still exist for a certain user
func (chain *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTxs []Transaction

	spentTXOs := make(map[string][]int)
	//iterate over the blockchain
	iter := chain.Iterator()

	for {
		block := iter.Next()
		//iterat over all transaction inside in a block
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			//iterat all tha output for that transaction
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil { //check if the output is inside our map
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				//check if the output can be unlocked by the address we're searching for
				if out.IsLockedWithKey(pubKeyHash) {
					//take all the transactions that can be unlocked
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			// if the transaction is not a Coinbase
			if tx.IsCoinBase() == false {
				//find other outputs that are references by inputs
				for _, in := range tx.Inputs {
					//check if we can unlock other outputs with the address, and than add them to the map
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspentTxs
}

//FindUTXO find all the unspent transaction outputs which correspond to an address.
func (chain *BlockChain) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := chain.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			//check if theoutput can be unlock by the address
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

//FindSpendableOutputs find all the unspent transaction output.
func (chain *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			//accumulated < amount: can't make a transaction if the user doesn't have enough tokens in their account
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts
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
