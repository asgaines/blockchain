package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

const blockchainFile = "storage.json"

type Blockchain []*Block

func InitBlockchain() Blockchain {
	f, err := os.Open(blockchainFile)
	if err != nil {
		genesis := NewBlock(&Block{}, []Transaction{}, 0)
		return Blockchain{genesis}
	}
	defer f.Close()

	var bc Blockchain

	decoder := json.NewDecoder(f)
	if err = decoder.Decode(&bc); err != nil {
		log.Fatal(err)
	}

	if !bc.IsSolid() {
		log.Fatal("Initialization failed due to broken chain")
	}

	return bc
}

func (bc Blockchain) AddBlock(block *Block) Blockchain {
	return append(bc, block)
}

func (bc *Blockchain) ToJSON() []byte {
	j, err := json.Marshal(bc)
	if err != nil {
		log.Fatal(err)
	}

	return j
}

func (bc Blockchain) IsSolid() bool {
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

func (bc Blockchain) LastLink() *Block {
	return bc[len(bc)-1]
}

func (bc Blockchain) TimeSinceLastLink() time.Duration {
	unixTsNano := bc.LastLink().Timestamp
	unixTsSec := unixTsNano / 1_000_000_000

	lastLinkTime := time.Unix(unixTsSec, 0)
	return time.Since(lastLinkTime)
}
