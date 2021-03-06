package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

//NewKeyPair create a Private and Public key
func NewKeyPair() (ecdsa.PrivateKey, []byte) {

	//define the type of the elliptic curve (in this cas is a 256)
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader) //generate the privateKey
	if err != nil {
		log.Panic(err)
	}
	//generate public key
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	return *privateKey, publicKey
}

//MakeWallet make the Wallet with the private a bublicKey.
func MakeWallet() *Wallet {
	privKey, pubKey := NewKeyPair()
	wallet := Wallet{PrivateKey: privKey, PublicKey: pubKey}
	return &wallet

}

// PublicKeyHash make a public key Hash.
func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}
	publicRipMD := hasher.Sum(nil)
	return publicRipMD
}

//CheckSum generate a checksum from a PublicKeyHash
func CheckSum(pubKeyHash []byte) []byte {
	firstHash := sha256.Sum256(pubKeyHash)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:checksumLength]

}

//Address generate an address for each wallet
func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)
	versionHash := append([]byte{version}, pubHash...)
	checkSum := CheckSum(versionHash)
	fullHash := append(versionHash, checkSum...)
	address := Base58Encode(fullHash)

	// fmt.Printf("pub key : %x\n", w.PublicKey)
	// fmt.Printf("pub hash: %x\n", pubHash)
	// fmt.Printf("address : %x\n", address)

	return address
}
