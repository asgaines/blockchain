package chain

import (
	"encoding/json"
	"log"

	"github.com/asgaines/blockchain/protogo/blockchain"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

type Chain struct {
	Pbc    *pb.Chain
	Hasher Hasher
}

func NewChain(hasher Hasher) *Chain {
	genesis := NewBlock(
		hasher,
		&Block{},
		[]*pb.Tx{},
		0,
		0,
		"",
	)

	chain := Chain{
		Pbc: &pb.Chain{
			Blocks: []*pb.Block{
				genesis.ToProto(),
			},
		},
		Hasher: hasher,
	}

	return &chain
}

func (bc *Chain) ToProto() *pb.Chain {
	return bc.Pbc
}

func (bc Chain) WithBlock(block *Block) *Chain {
	return &Chain{
		Pbc: &blockchain.Chain{
			Blocks: append(bc.Pbc.Blocks, block.ToProto()),
		},
	}
}

func (bc Chain) IsSolid() bool {
	if len(bc.Pbc.Blocks) <= 0 {
		return false
	}

	for i, block := range bc.Pbc.Blocks[1:] {
		prev := bc.Pbc.Blocks[i]

		prevhash := bc.Hasher.Hash((*Block)(prev))
		if prevhash != block.Prevhash ||
			prevhash != prev.Hash ||
			block.Hash != bc.Hasher.Hash((*Block)(block)) ||
			float64(block.Hash) > block.Target {
			return false
		}
	}

	return true
}

func (bc Chain) BlockByIdx(idx int) *Block {
	return (*Block)(bc.Pbc.Blocks[idx])
}

func (bc Chain) LastLink() *Block {
	return bc.BlockByIdx(len(bc.Pbc.Blocks) - 1)
}

func (bc Chain) Length() int {
	return len(bc.Pbc.Blocks)
}

func (bc Chain) ToJSON() []byte {
	j, err := json.Marshal(bc)
	if err != nil {
		log.Fatal(err)
	}

	return j
}
