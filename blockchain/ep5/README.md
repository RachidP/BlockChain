# BlockChain
//print blockchain
go run main.go printchain

// create a blockchain command with the address john inside of it
go run main.go createblockchain -address "John"

//check balance
go run main.go getbalance -address "John"

//send account from John to Fred
go run main.go send -from "John" -to "Fred" -amount 50


//create wallet
go run main.go createwallet

//list addresses
go run main.go listaddresses
