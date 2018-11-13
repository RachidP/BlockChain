package blockchain

// BlockChain rappresent a BlockChain
type BlockChain struct {
	Blocks []*Block
}

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

// InitialBlockChain build our initial BlockChian.
func InitialBlockChain() *BlockChain {
	firstBlock := Genesis()
	return &BlockChain{[]*Block{firstBlock}}
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

// AddBlock add a Block to the BlockChain
func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.Blocks[len(chain.Blocks)-1] //get previous Block in our BlockChain
	currBlock := CreateBlock(data, prevBlock.Hash) //create a Block
	chain.Blocks = append(chain.Blocks, currBlock) //add the current block to the chain
}
