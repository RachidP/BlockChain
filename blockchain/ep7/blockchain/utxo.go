package blockchain

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/dgraph-io/badger"
)

var (
	utxoPrefix   = []byte("utxo-") //is a prefix for the keys in the DB for separate data i(nto badger DB) because badger doesn't have tables
	prefixLength = len(utxoPrefix)
)

//UTXOSet is the main structure for the unspent transaction outputs
type UTXOSet struct {
	Blockchain *BlockChain
}

//FindUTXO FindUnspentTransactions
//Unspent transactions are transactions that have output wich are not referenced by other inputs
//that's means that are tokens still exist for a certain user
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput

	db := u.Blockchain.Database

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			v, err := item.Value()
			HandleErr(err)
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	HandleErr(err)

	return UTXOs
}

// FindSpendableOutputs create and send transactions inside the blockchain.
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.Database
	//iterate over the db
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			k := item.Key()        //get the key inside the DB
			v, err := item.Value() // get the value from the db
			HandleErr(err)
			k = bytes.TrimPrefix(k, utxoPrefix) //trim the prefix from the key
			txID := hex.EncodeToString(k)       //encode the key into a string
			outs := DeserializeOutputs(v)       //deserialize it into output struct
			//iterate over the outputs
			for outIdx, out := range outs.Outputs {
				//check if the output has been looked with the pubKeyHash
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outIdx)
				}
			}
		}
		return nil
	})
	HandleErr(err)
	return accumulated, unspentOuts
}

func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.Database
	counter := 0

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			counter++
		}

		return nil
	})

	HandleErr(err)

	return counter
}

//Reindex
func (u UTXOSet) Reindex() {
	db := u.Blockchain.Database //allias  the db

	u.DeleteByPrefix(utxoPrefix)

	UTXO := u.Blockchain.FindUTXO()

	err := db.Update(func(txn *badger.Txn) error {
		for txId, outs := range UTXO {
			key, err := hex.DecodeString(txId)
			if err != nil {
				return err
			}
			key = append(utxoPrefix, key...)
			//push into db
			err = txn.Set(key, outs.Serialize())
			HandleErr(err)
		}

		return nil
	})
	HandleErr(err)
}

//Update the UTXOset inside the db by taken the block.
func (u *UTXOSet) Update(block *Block) {
	db := u.Blockchain.Database

	err := db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					updatedOuts := TxOutputs{}
					inID := append(utxoPrefix, in.ID...)
					item, err := txn.Get(inID)
					HandleErr(err)
					v, err := item.Value()
					HandleErr(err)

					outs := DeserializeOutputs(v)

					for outIdx, out := range outs.Outputs {
						if outIdx != in.Out {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						if err := txn.Delete(inID); err != nil {
							log.Panic(err)
						}

					} else {
						if err := txn.Set(inID, updatedOuts.Serialize()); err != nil {
							log.Panic(err)
						}
					}
				}
			}

			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			txID := append(utxoPrefix, tx.ID...)
			if err := txn.Set(txID, newOutputs.Serialize()); err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	HandleErr(err)
}

//DeleteByPrefix go throw the DB and delete in bulk the prefix keys from the DB.
func (u *UTXOSet) DeleteByPrefix(prefix []byte) {
	//access to the db and delete the key if no errors
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := u.Blockchain.Database.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}
	//set the optimal ammout of keys that can be deleted in one batch delete with badger DB
	collectSize := 100000

	u.Blockchain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize) //collect the keys that they will be deleted
		keysCollected := 0

		//iterate over all the keys that contain the "prefix"
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().KeyCopy(nil)              //copy the key itself
			keysForDelete = append(keysForDelete, key) //add the key to delete into keysForDelete
			keysCollected++
			//if the the number the keys to delete is 100000 we can delete all theese keys
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					log.Panic(err)
				}
				//reset the data for the next for cycle
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		// if the are others key to delete that are less than 10000 keys, delete them
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}
