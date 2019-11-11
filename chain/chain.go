package chain

import (
	"encoding/json"
	"log"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

type Chain pb.Chain

func NewChain() *Chain {
	genesis := NewBlock(
		&Block{},
		[]*pb.Tx{},
		0,
		0,
		"",
	)

	chain := Chain{
		Blocks: []*pb.Block{
			genesis.ToProto(),
		},
	}

	return &chain
}

func (bc *Chain) ToProto() *pb.Chain {
	return (*pb.Chain)(bc)
}

func (bc Chain) AddBlock(block *Block) *Chain {
	bc.Blocks = append(bc.Blocks, block.ToProto())
	return &bc
}

func (bc Chain) IsSolid() bool {
	if len(bc.Blocks) <= 0 {
		return false
	}

	for i, block := range bc.Blocks[1:] {
		prev := bc.Blocks[i]

		prevhash := (*Block)(prev).makeHash()
		if prevhash != block.Prevhash ||
			prevhash != prev.Hash ||
			block.Hash != (*Block)(block).makeHash() ||
			block.Hash > block.Target {
			return false
		}
	}

	return true
}

func (bc Chain) BlockByIdx(idx int) *Block {
	return (*Block)(bc.Blocks[idx])
}

func (bc Chain) LastLink() *Block {
	return bc.BlockByIdx(len(bc.Blocks) - 1)
}

func (bc Chain) Length() int {
	return len(bc.Blocks)
}

func (bc Chain) ToJSON() []byte {
	j, err := json.Marshal(bc)
	if err != nil {
		log.Fatal(err)
	}

	return j
}
