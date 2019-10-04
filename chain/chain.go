package chain

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/asgaines/blockchain/transactions"
)

const BlockchainFile = "storage.json"

type Chain []*Block

func InitBlockchain() Chain {
	f, err := os.Open(BlockchainFile)
	if err != nil {
		genesis := NewBlock(&Block{}, []transactions.Tx{}, 0, 0)
		return Chain{genesis}
	}
	defer f.Close()

	var bc Chain

	decoder := json.NewDecoder(f)
	if err = decoder.Decode(&bc); err != nil {
		log.Fatal(err)
	}

	if !bc.IsSolid() {
		log.Fatal("Initialization failed due to broken chain")
	}

	return bc
}

func (bc Chain) AddBlock(block *Block) Chain {
	return append(bc, block)
}

func (bc Chain) ToJSON() []byte {
	j, err := json.Marshal(bc)
	if err != nil {
		log.Fatal(err)
	}

	return j
}

func (bc Chain) IsSolid() bool {
	for i, block := range bc[1:] {
		prev := bc[i]

		prevhash := prev.makeHash()
		if prevhash != block.PrevHash ||
			prevhash != prev.Hash ||
			block.Hash != block.makeHash() {
			return false
		}
	}

	return true
}

func (bc Chain) LastLink() *Block {
	return bc[len(bc)-1]
}

func (bc Chain) TimeSinceLastLink() time.Duration {
	unixTsNano := bc.LastLink().Timestamp
	unixTsSec := unixTsNano / 1_000_000_000

	lastLinkTime := time.Unix(unixTsSec, 0)
	return time.Since(lastLinkTime)
}
