package blockchain

import "../labrpc"
import "sync"
import "time"
import (
    "fmt"
    "strings"
    "crypto/sha1"
    "encoding/base64"
)

//
// A Go object implementing a single Blockchain peer.
//
type Blockchain struct {
    mu             sync.Mutex          // Lock to protect shared access to this peer's state
    peers          []*labrpc.ClientEnd // RPC end points of all peers
    me             int                 // this peer's index into peers[]

    chains         []Block         // This represent the entire block chain
    commandChannel chan string     // Channel to listen to broadcast and download command
    difficulty     int             // Difficulty level of the blockchain
}

// the tester calls Kill() when a Blockchain instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (bc *Blockchain) Kill() {
    // Your code here, if desired.
}

//
// Function to mine first ever block for every peer's blockchain.
// The first block is genesis block.
//
func (bc *Blockchain) createGenesisBlock() {

    // Your code to mine the genesis block.
    // It should be the first block mined by all blockchain peers on starting up.
    block := Block{GENESIS_HASH, "", "first block", 0, time.Now(), 0}
    bc.chains = append(bc.chains, block)
}

//
// Tester will invoke CreateNewBlock API on a peer for mining a new block.
// Tester will provide the Block data.
//
type CreateNewBlockArgs struct {
    BlockData    string
}

type CreateNewBlockReply struct {
    // Your code here, if desired
    success bool
}

func (bc *Blockchain) CreateNewBlock(args *CreateNewBlockArgs, reply *CreateNewBlockReply) {
    block := Block{}
    block.data = args.BlockData
    block = bc.mine(block)
    fmt.Printf(block.hash)
    reply.success = true
}

//
// A blockchain peer will invoke DownloadBlockchain RPC on another blockchain
// peer to download its blockchain.
//
type DownloadBlockchainArgs struct {
    // Your code here
}

type DownloadBlockchainReply struct {
    // Your code here
}

func (bc *Blockchain) DownloadBlockchain(args *DownloadBlockchainArgs, reply *DownloadBlockchainReply) {
    // Your code here
}

//
// Invoked by tester to instruct a blockchain peer to update it's blockchain
// after reconnection.
//
func (bc *Blockchain) DownloadBlockchainFromPeers() {

    // Your code here
}

//
// Invoked by tester to instruct a blockchain peer to broadcast its pending created
// block to other peers. Invoked after mining new block on this peer.
// Unless all other peers agree on this pending block,
// it cant be added to the final blockchain of this peer.
//
func (bc *Blockchain) BroadcastNewBlock() {
    // Your code here
}

//
// A blockchain peer will invoke AddBlock RPC on another blockchain peer
// to request for adding its newly mined block.
//
type AddBlockArgs struct {
    // Your code here
}

type AddBlockReply struct {
    // Your code here
}

func (bc *Blockchain) AddBlock(args *AddBlockArgs, reply *AddBlockReply) {
    // Your code here
}

//
// Invoked by tester to get blockchain length and last block's hash.
//
func (bc *Blockchain) GetState() (int, string) {
    size := len(bc.chains)

    return size, bc.chains[size-1].hash
}

func (bc *Blockchain) mine(block Block) (Block) {
    block.nonce = 0
    requiredPrefix := strings.Repeat("0", bc.difficulty)

    block.previousHash = bc.chains[len(bc.chains) - 1].hash

    for {
        blockBinary := []byte(fmt.Sprintf("%s%s%d", block.data, block.previousHash, block.nonce))

        hasher := sha1.New()
        hasher.Write(blockBinary)
        hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

        fmt.Printf("The current hash is %s\n", hash)

        if strings.HasPrefix(hash, requiredPrefix) {
            block.hash = hash
            break
        }

        block.nonce++
    }

    return block
}
func (bc *Blockchain) Listen() {
    for cmd := range bc.commandChannel {
        if cmd == "Broadcast" {
            bc.commandChannel <- "Done"
        } else if cmd == "DownloadBlockchain"{
            bc.commandChannel <- "Done"
        }
    }
}
//
// the service or tester wants to create a Blockchain server. the ports
// of all the Blockchain servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
// Command channel is used by tester to instruct Broadcast and DownloadBlockchain
// commands to this blockchain peer.
//
func Make(peers []*labrpc.ClientEnd, me int, difficulty int, command chan string) *Blockchain {

    bc := &Blockchain{}
    bc.peers = peers
    bc.me = me

    // Your initialization code here
    bc.difficulty = difficulty
    bc.commandChannel = command
    bc.chains = [] Block{}
    bc.createGenesisBlock()
    go bc.Listen()
    return bc
}
