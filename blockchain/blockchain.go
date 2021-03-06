package blockchain

import "../labrpc"
import "sync"
import "time"
import (
    "fmt"
    "strings"
    "encoding/base64"
    "crypto/sha256"
)

//
// A Go object implementing a single Blockchain peer.
//
type Blockchain struct {
    mu             sync.Mutex          // Lock to protect shared access to this peer's state
    peers          []*labrpc.ClientEnd // RPC end points of all peers
    me             int                 // this peer's Index into peers[]

    chains         []Block         // This represent the entire block chain
    commandChannel chan string     // Channel to listen to broadcast and download command
    difficulty     int             // Difficulty level of the blockchain
    newBlock       Block           // The newly mined block that has yet to be broadcasted
    prefix string                  // a prefix that every hash must contain
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
    bc.mu.Lock()
    bc.chains = append(bc.chains, block)
    bc.mu.Unlock()
}

//
// Tester will invoke CreateNewBlock API on a peer for mining a new block.
// Tester will provide the Block Data.
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
    block.Data = args.BlockData
    block = bc.mine(block)
    fmt.Printf("newly mined block by peer %d, has hash %s, timestamp %s, and index %d\n", bc.me, block.Hash, block.Timestamp, block.Index)
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
    Chains []Block
}

//
// Send addBlock request to a particular peer
func (bc *Blockchain) sendDownloadBlockchain(server int, args *DownloadBlockchainArgs, reply *DownloadBlockchainReply) bool {
    ok := bc.peers[server].Call("Blockchain.DownloadBlockchain", args, reply)
    return ok
}

func (bc *Blockchain) DownloadBlockchain(args *DownloadBlockchainArgs, reply *DownloadBlockchainReply) {
    // Your code here
    bc.mu.Lock()
    reply.Chains = bc.chains
    bc.mu.Unlock()
}

//
// Invoked by tester to instruct a blockchain peer to update it's blockchain
// after reconnection.
//
func (bc *Blockchain) DownloadBlockchainFromPeers() {

    // Your code here
    args := DownloadBlockchainArgs{}
    type ResponseMsg struct {
        DownloadBlockchainReply
        IsOk      bool
        PeerIndex int
    }

    responseChan := make(chan ResponseMsg)
    // Send request concurrently
    for i, _ := range bc.peers {
        if i == bc.me {
            continue
        }

        go func(peerIndex int) {
            resp := DownloadBlockchainReply{}
            ok := bc.sendDownloadBlockchain(peerIndex, &args, &resp)
            responseChan <- ResponseMsg{
                resp,
                ok,
                peerIndex,
            }
        }(i)
    }

    // collect response
    bc.mu.Lock()
    maxChain := bc.chains
    maxLen := len(bc.chains)
    minTimestamp := bc.chains[maxLen-1].Timestamp
    bc.mu.Unlock()
    totalCount := len(bc.peers)
    currentCount := 1
    for resp := range responseChan {
        currentCount++
        currentChain := resp.Chains
        currentLen := len(currentChain)
        if resp.IsOk == false || currentLen == 0 {
            continue
        }
        fmt.Printf("current chain length is %d from peer %d\n", currentLen, resp.PeerIndex)
        currentTime := currentChain[currentLen-1].Timestamp
        if currentLen > maxLen {
            maxLen = currentLen
            maxChain = currentChain
            minTimestamp = currentTime
        } else if currentLen == maxLen && currentTime.Before(minTimestamp) {
            maxChain = currentChain
            minTimestamp = currentTime
        }

        if currentCount == totalCount {
            bc.mu.Lock()
            bc.chains = maxChain
            bc.mu.Unlock()
            bc.commandChannel <- "Done"
            return
        }
    }
}

//
// Invoked by tester to instruct a blockchain peer to broadcast its pending created
// block to other peers. Invoked after mining new block on this peer.
// Unless all other peers agree on this pending block,
// it cant be added to the final blockchain of this peer.
//
func (bc *Blockchain) BroadcastNewBlock() {
    // Your code here
    bc.mu.Lock()
    args := AddBlockArgs{
        Index:  bc.newBlock.Index,
        PreviousHash: bc.newBlock.PreviousHash,
        Timestamp: bc.newBlock.Timestamp,
        Hash: bc.newBlock.Hash,
        Data: bc.newBlock.Data,
        Nonce: bc.newBlock.Nonce,
    }
    bc.mu.Unlock()
    type ResponseMsg struct {
        AddBlockReply
        IsOk      bool
        PeerIndex int
    }

    responseChan := make(chan ResponseMsg)
    // send requests concurrently
    for i, _ := range bc.peers {
        if i == bc.me {
            continue
        }

        go func(peerIndex int) {
            resp := AddBlockReply{}
            resp.Approved = true
            ok := bc.sendAddBlock(peerIndex, &args, &resp)
            responseChan <- ResponseMsg{
                resp,
                ok,
                peerIndex,
            }
        }(i)
    }

    // Collect response
    totalCount := len(bc.peers)
    currentCount := 1
    for resp := range responseChan {
        if resp.AddBlockReply.Approved == false {
            bc.newBlock = Block{}
            go bc.DownloadBlockchainFromPeers()
            return
        } else {
            currentCount++
            if currentCount == totalCount {
                fmt.Printf("node %d received all approval\n", bc.me)
                bc.mu.Lock()
                if bc.newBlock.Index == len(bc.chains) {
                    bc.chains = append(bc.chains, bc.newBlock)
                    bc.newBlock = Block{}
                    bc.mu.Unlock()
                } else {
                    // This happens when the node accepts another peer's mined block
                    bc.newBlock = Block{}
                    bc.mu.Unlock()
                    go bc.DownloadBlockchainFromPeers()
                    return
                }

                bc.commandChannel <- "Done"
                return
            }
        }
    }
}

//
// Send addBlock request to a particular peer
func (bc *Blockchain) sendAddBlock(server int, args *AddBlockArgs, reply *AddBlockReply) bool {
    ok := bc.peers[server].Call("Blockchain.AddBlock", args, reply)
    return ok
}

//
// A blockchain peer will invoke AddBlock RPC on another blockchain peer
// to request for adding its newly mined block.
//
type AddBlockArgs struct {
    Index           int // index of the new block
    PreviousHash    string // previous hash
    Timestamp       time.Time // timestamp for the new block
    Hash            string // hash for the new block
    Data            string // block data
    Nonce           int // block nonce
}

type AddBlockReply struct {
    // Your code here
    Approved bool // if the block is approved or not
}

func (bc *Blockchain) AddBlock(args *AddBlockArgs, reply *AddBlockReply) {
    // Your code here
    bc.mu.Lock()
    defer bc.mu.Unlock()
    length := len(bc.chains)
	newBlock := Block{
		args.Hash,
		args.PreviousHash,
		args.Data,
		args.Index,
		args.Timestamp,
		args.Nonce,
	}

	// Check the all conditions to approve a block
	isValid := args.Index == length &&
		args.PreviousHash == bc.chains[length-1].Hash &&
		strings.HasPrefix(args.Hash, bc.prefix) &&
		newBlock.Hash == bc.getBlockHash(newBlock)

    if isValid {
        bc.chains = append(bc.chains, newBlock)
        reply.Approved = true
    } else {
        reply.Approved = false
    }
}

//
// Invoked by tester to get blockchain length and last block's Hash.
//
func (bc *Blockchain) GetState() (int, string) {
    bc.mu.Lock()
    size := len(bc.chains)
    hash := bc.chains[size-1].Hash
    bc.mu.Unlock()

    return size, hash
}

func (bc *Blockchain) getBlockHash(block Block) string {
	blockBinary := []byte(
		fmt.Sprintf("%s%s%d%d%d",
			block.Data,
			block.PreviousHash,
			block.Nonce),
	)

	hasher := sha256.New()
	hasher.Write(blockBinary)
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return hash
}

func (bc *Blockchain) mine(block Block) (Block) {
    bc.mu.Lock()
    defer bc.mu.Unlock()
    block.Nonce = 0

    block.PreviousHash = bc.chains[len(bc.chains) - 1].Hash

    for {
    	hash := bc.getBlockHash(block)

        if strings.HasPrefix(hash, bc.prefix) {
            block.Hash = hash
            block.Index = len(bc.chains)
            block.Timestamp = time.Now()
            bc.newBlock = block
            break
        }

        block.Nonce++
    }

    return block
}

func (bc *Blockchain) Listen() {
    for cmd := range bc.commandChannel {
        if cmd == "Broadcast" {
            fmt.Printf("received broadcast command from tester\n")
            if bc.newBlock.Hash != "" {
                go bc.BroadcastNewBlock()
            }
        } else if cmd == "DownloadBlockchain"{
            fmt.Printf("received download command from tester\n")
            go bc.DownloadBlockchainFromPeers()
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
	bc.prefix = strings.Repeat("0", bc.difficulty)
    bc.createGenesisBlock()
    go bc.Listen()
    return bc
}
