package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/RachidP/BlockChain/blockchain"
)

func main() {

	defer os.Exit(0)
	chain := blockchain.InitialBlockChain()

	defer chain.Database.Close()
	cli := CommandLine{chain}
	cli.run()
}

//CommandLine Allow the user to pass a new Block from command line.
type CommandLine struct {
	blockchain *blockchain.BlockChain
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("add -block BLOCK_DATA - add a block to the chain")
	fmt.Println("print -Prints the blocks in the chain")

}

// ValidateArgs validate the args passed from the command line.
func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit() // is a safe exit it shotdown the goroutines and it safe for working with the data in the db
	}

}

func (cli *CommandLine) addBlock(data string) {

	cli.blockchain.AddBlock(data)
	fmt.Println("Added Block!")
}
func (cli *CommandLine) printChain() {
	iter := cli.blockchain.Iterator() //get iterator
	for {
		block := iter.Next()

		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		// if we are at the Genesis block
		if len(block.PrevHash) == 0 {
			break
		}

	}

}

func (cli *CommandLine) run() {
	cli.validateArgs()
	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")

	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		blockchain.HandleErr(err)

	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.HandleErr(err)
	default:
		cli.printUsage()
		runtime.Goexit()
	}
	//if addBlockCmd has been parsed
	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()

		}
		cli.addBlock(*addBlockData)

	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

}
