package blockchain

import "time"
const GENESIS_HASH = "816534932c2b7154836da6afc367695e6337db8a921823784c14378abed4f7d7"

type Block struct {

    // Your code here for what goes into a blockchain's block.
	Hash         string    //the Hash value for the block
	PreviousHash string    // the previous Hash
	Data         string    // Data for the current block
	Index        int       // Index of this block in the chain
	Timestamp    time.Time // Timestamp of the block
	Nonce        int       // Nonce of the block
}