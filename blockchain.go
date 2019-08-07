package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

const blockchainFile = "storage.json"

// 1
type Blockchain []*Block

func InitBlockchain() Blockchain {
	f, err := os.Open(blockchainFile)
	if err != nil {
		genesis := NewBlock(&Block{}, LilBits{})
		return Blockchain{genesis}
	}
	defer f.Close()

	var bc Blockchain

	decoder := json.NewDecoder(f)
	if err = decoder.Decode(&bc); err != nil {
		log.Fatal(err)
	}

	bc.Verify()

	return bc
}

func (bc *Blockchain) addBlock(payload LilBits) {
	block := NewBlock((*bc)[len(*bc)-1], payload)
	*bc = append(*bc, block)

	bc.Propagate()
}

func (bc Blockchain) Propagate() {
	bc.Verify()

	if err := ioutil.WriteFile(blockchainFile, bc.ToJSON(), 0644); err != nil {
		log.Fatal(err)
	}
}

func (bc *Blockchain) ToJSON() []byte {
	j, err := json.Marshal(bc)
	if err != nil {
		log.Fatal(err)
	}

	return j
}

func (bc Blockchain) Verify() {
	for i, block := range bc[1:] {
		prev := bc[i]

		prevhash := prev.makeHash()
		if prevhash != block.PrevHash ||
			prevhash != prev.Hash ||
			block.Hash != block.makeHash() {
			log.Fatal("Block is not valid!")
		}
	}
}
