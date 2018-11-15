package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

// Transaction describe a transaction
type Transaction struct {
	ID      []byte //it's a Hash
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxOutput struct {
	Value  int    //the values are in tokens
	Pubkey string // is the public key for unlock the token (inside the Value)
}

//TxInput is a references to previous outputs
type TxInput struct {
	ID  []byte //reference the transaction that the output is inside
	Out int    //index of the output appears (ID)
	Sig string
}

// CoinbaseTx make the first transaction for the block Genesis
func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		//make a new data
		data = fmt.Sprintf("Coins to %s", to)
	}

	//define the Transaction input and output for this Coinbase
	txin := TxInput{
		ID:  []byte{}, //is nill because is referencing no output
		Out: -1,       //-1 because is referencing no output
		Sig: data,
	}
	txout := TxOutput{
		Value:  100, //has the reward is 100 tokens
		Pubkey: to,
	}

	//create the transaction
	tx := Transaction{ID: nil,
		Inputs:  []TxInput{txin},
		Outputs: []TxOutput{txout},
	}
	tx.SetID()
	return &tx

}

//SetId make the Hash  for the ID transaction
func (tx *Transaction) SetID() {

	var encoded bytes.Buffer
	var hash [32]byte
	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	HandleErr(err)
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]

}

//IsCoinBase determine whether of not a transaction is a coinbase transaction.
func (tx *Transaction) IsCoinBase() bool {
	//the coinbase only has one input
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

//CanUnlock Unlock the data from the input.
//that means that the account owns the data.
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data

}

//CanBeUnlocked Unlock the data from the output.
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.Pubkey == data

}

//NewTransaction
func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := chain.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		HandleErr(err)
		//create a input for each unspent output
		for _, out := range outs {
			input := TxInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	//create the output for the transaction
	outputs = append(outputs, TxOutput{amount, to})
	//the ammount from that the user has is  greater than the user is trying to send
	if acc > amount {
		//create a second output
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}
