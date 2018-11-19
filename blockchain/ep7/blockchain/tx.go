package blockchain

import (
	"bytes"
	"encoding/gob"

	"github.com/RachidP/BlockChain/wallet"
)

type TxOutput struct {
	Value      int    //the values are the amount of tokens inside of it
	PubKeyHash []byte // is the public key for unlock the token (inside the Value)
}

//TxOutputs identify transactions outputs, and sort them by unspent output
type TxOutputs struct {
	Outputs []TxOutput
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

//Serialize encode the strture TxOutputs into []byte
func (outs TxOutputs) Serialize() []byte {
	var buffer bytes.Buffer

	encode := gob.NewEncoder(&buffer)
	err := encode.Encode(outs)
	HandleErr(err)

	return buffer.Bytes()
}

//DeserializeOutputs decode the byte into structure TxOutputs
func DeserializeOutputs(data []byte) TxOutputs {
	var outputs TxOutputs

	decode := gob.NewDecoder(bytes.NewReader(data))
	err := decode.Decode(&outputs)
	HandleErr(err)

	return outputs
}
