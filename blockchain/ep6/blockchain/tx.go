package blockchain

import (
	"bytes"

	"github.com/RachidP/BlockChain/wallet"
)

type TxOutput struct {
	Value      int    //the values are the amount of tokens inside of it
	PubKeyHash []byte // is the public key for unlock the token (inside the Value)
}

//TxInput is a references to previous outputs
type TxInput struct {
	ID        []byte //reference the transaction that the output is inside
	Out       int    //index of the output appears (ID)
	Signature []byte
	PubKey    []byte
}

//UsesKey
func (in *TxInput) UsesKey(pubkeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)
	return bytes.Compare(lockingHash, pubkeyHash) == 0
}

//Lock lock the output.
func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

//IsLockedWithKey check if the output is locked.
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

//NewTXOutput
func NewTXOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}
