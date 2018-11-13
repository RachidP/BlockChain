package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

//Block contain the basic data for Blockchain.
type Block struct {
	Hash     []byte // the hash who rappresent the block itself
	Data     []byte // data inside this block (document, images)
	PrevHash []byte // rappresent the last block hash, allows to link block together
	Nonce    int    // is used to derived the hash(which met the target )
}

// Genesis create the first Inizial block in the blockChian.
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

// CreateBlock Create the current block.
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{
		Hash:     []byte{},
		Data:     []byte(data),
		PrevHash: prevHash,
		Nonce:    0,
	}
	pow := NewProof(block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash
	return block

}

// Serialize encode the data from a block to a []bytes
// this method help us to work with the DB because our DB works only with an arrays of bytes
func (b *Block) Serialize() []byte {

	var res bytes.Buffer

	encoder := gob.NewEncoder(&res) //make a new encoder

	err := encoder.Encode(b) // encode our block
	HandleErr(err)
	//return bytes rappresentation of our block
	return res.Bytes()
}

// Deserialize decode the data from []bytes to a Block
func Deserialize(data []byte) *Block {
	var block Block

	reader := bytes.NewReader(data)   //make new reader
	decoder := gob.NewDecoder(reader) //make new decoder
	err := decoder.Decode(block)
	HandleErr(err)
	return &block

}

// HandleErr is halper to handle the errors
func HandleErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}
