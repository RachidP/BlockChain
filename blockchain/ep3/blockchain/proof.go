package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

//	TO DO
//1)Take the data from the block
//2)create a counter (nonce) which starts at 0
//3)create a hash of the data plus the counter
//4)check the hash too see if it meets a set of requirements
//Requirements:
//The first few bytes must contans 0s

// Difficulty in this example is costant, but in real blockchain you need to
//implemenet an algorithm that would slowly increment this difficulty
//over a large period of time
const Difficulty = 20

// ProofOfWork is the struct
type ProofOfWork struct {
	Block  *Block   //Block is the block inside the blockchain
	Target *big.Int //Target is a number that rappresents the requirement wich dirived from Difficulty
}

//NewProof Initialize a new ProofOfWork, by taking the data from the block (1)
func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	//256 is the number of bytes inside one of our hashes
	//left shift the target of 256bytes-Difficulty
	target.Lsh(target, uint(256-Difficulty))
	pow := &ProofOfWork{Block: b, Target: target}
	return pow
}

// InitData Derive the hash based on the previous Hash and the current data.
// is it like DeriveHash but with our HashFunction
// Create our counter or nonce (2)
func (pow *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.Data,
			ToHex(int64(nonce)),
			ToHex(int64(Difficulty)),
		},
		[]byte{},
	)
	return data

}

// ToHex is utility function that convert int64 to a []byte organized in BigEndian form
func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	//take our nummber (num) and decode into bytes
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// Run Create the Hash from the counter plus the data and
// then check if the hash meets a set of requirements
func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte
	nonce := 0
	//prepare the data, then hash it into a sha-256 format than
	//convert that hash into a big integer and
	//compare that big integer with our target big integer
	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)

		intHash.SetBytes(hash[:]) //convert our hash into a bigInteger
		//compare our hash with the target and if the come back with -1 we break
		//out of loop because that mean that our hash is less than the target we're looking for
		//which means that we've actually signed the block
		if intHash.Cmp(pow.Target) == -1 {
			break
		}

		nonce++
	}
	fmt.Println()
	return nonce, hash[:]
}

// Validate After Run() function will have nonce wich allow us to derive the hash, which
//met the target that we wanted, then we'll able to simply run that cycle one more time to show
//that this hash is valid
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitData(pow.Block.Nonce) //take the data

	hash := sha256.Sum256(data) //convert data into hash

	intHash.SetBytes(hash[:]) //convert hash into bigInt a put it into intHash variable

	//check ig the block itself is valid
	return intHash.Cmp(pow.Target) == -1
}
