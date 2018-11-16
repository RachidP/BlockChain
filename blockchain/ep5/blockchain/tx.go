package blockchain

type TxOutput struct {
	Value  int    //the values are in tokens
	Pubkey string // is the public key for unlock the token (inside the Value)
}

//TxInput is a references to previous outputs
type TxInput struct {
	ID  []byte //reference the transaction that the output is inside
	Out int    //index of the output appears (ID)
	Sig string
}

//CanUnlock Unlock the data from the input.
//that means that the account owns the data.
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data

}

//CanBeUnlocked Unlock the data from the output.
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.Pubkey == data

}
