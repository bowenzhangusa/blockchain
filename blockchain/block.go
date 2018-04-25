package blockchain

import "time"
const GENESIS_HASH = "816534932c2b7154836da6afc367695e6337db8a921823784c14378abed4f7d7"

type Block struct {

    // Your code here for what goes into a blockchain's block.
	hash string //the hash value for the block
	previousHash string // the previous hash
	data string // data for the current block
	index int // index of this block in the chain
	timestamp time.Time // timestamp of the block
	nonce int // nonce of the block
}