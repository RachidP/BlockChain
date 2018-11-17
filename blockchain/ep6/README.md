# BlockChain
//print blockchain
go run main.go printchain





//delete all the files inside tmp/blocks

//1) create 2 wallet
go run main.go createwallet
go run main.go createwallet

// 2) create a blockchain command with one of the prev addresses 
go run main.go createblockchain -address 1231SWoo8eYKpwaQ1ynLKASuV5YS1R1AaD

//3) print blockchain
go run main.go printchain


//4) send account from John to Fred
go run main.go send -to 1H9iprVG6i5DdZiLcoTYokWEjvxTS27gGT -from 13pBbswb1SP4pqTVFCyNPardNQikJD74VK  -amount 30

//5) print blockchain
go run main.go printchain

//check balance
go run main.go getbalance -address 1H9iprVG6i5DdZiLcoTYokWEjvxTS27gGT

//list addresses
go run main.go listaddresses
