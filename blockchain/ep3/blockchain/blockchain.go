package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger"
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

const (
	dbPath = "./tmp/blocks"
)

// InitialBlockChain build our initial BlockChian.
func InitialBlockChain() *BlockChain {
	var lastHash []byte

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
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			//no existing blockchain in our db
			fmt.Println("No existing blockchain found")

			genesis := Genesis() //make genesis block
			fmt.Println("Genesis proved")

			//uses the genesis hash as the key of the genesis hash, serialize the block and put into db
			err := txn.Set(genesis.Hash, genesis.Serialize())
			HandleErr(err)
			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash

			return err

		} else {
			// if the blockchain has  already been stored into db
			item, err := txn.Get([]byte("lh"))
			HandleErr(err)
			lastHash, err = item.Value()
			return err

		}

	})

	HandleErr(err)
	blockchain := BlockChain{LastHash: lastHash, Database: db}
	return &blockchain
}

// AddBlock add a Block to the BlockChain
func (chain *BlockChain) AddBlock(data string) {

	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		HandleErr(err)
		lastHash, err = item.Value()
		return err
	})
	HandleErr(err)

	newBlock := CreateBlock(data, lastHash)

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
