package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// BlockChain rappresent a BlockChain
type BlockChain struct {
	blocks []*Block
}

//Block contain the basic data for Blockchain.
type Block struct {
	Hash     []byte // the hash who rappresent the block itself
	Data     []byte // data inside this block (document, images)
	PrevHash []byte // rappresent the last block hash, allows to link block together
}

// Genesis create the first Inizial block in the blockChian.
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

// InitialBlockChain build our initial BlockChian.
func InitialBlockChain() *BlockChain {
	firstBlock := Genesis()
	return &BlockChain{[]*Block{firstBlock}}
}

// DeriveHash Derive the hash based on the previous Hash and the current data.
func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{}) //join current hash with prevHash
	hash := sha256.Sum256(info)                                //create hash
	b.Hash = hash[:]                                           //add the created Hash to the hash of the current block
}

// CreateBlock Create the current block.
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{
		Hash:     []byte{},
		Data:     []byte(data),
		PrevHash: prevHash,
	}
	block.DeriveHash()
	return block

}

// AddBlock add a Block to the BlockChain
func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.blocks[len(chain.blocks)-1] //get previous Block in our BlockChain
	currBlock := CreateBlock(data, prevBlock.Hash) //create a Block
	chain.blocks = append(chain.blocks, currBlock) //add the current block to the chain
}

func main() {
	chain := InitialBlockChain()

	chain.AddBlock("First Block After Genesis")
	chain.AddBlock("Second Block After Genesis")
	chain.AddBlock("Third Block After Genesis")

	for _, block := range chain.blocks {
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n\n\n", block.Hash)

	}

}
