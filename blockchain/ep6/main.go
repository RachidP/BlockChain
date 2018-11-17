package main

import (
	"os"

	"github.com/RachidP/BlockChain/cli"
)

func main() {
	defer os.Exit(0)
	cmd := cli.CommandLine{}
	cmd.Run()
}
