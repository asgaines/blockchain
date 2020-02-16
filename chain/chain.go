package chain

import (
	"encoding/json"
	"log"

	"github.com/asgaines/blockchain/protogo/blockchain"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

type Chain struct {
	Pbc *pb.Chain
}

func NewChain(hasher Hasher) *Chain {
	genesis := NewBlock(
		hasher,
		&Block{},
		[]*pb.Tx{},
		0,
		[]byte{},
		"",
	)

	chain := Chain{
		Pbc: &pb.Chain{
			Blocks: []*pb.Block{
				genesis.ToProto(),
			},
		},
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

func (bc *Chain) GetCreditFor(pubkey string) float64 {
	credit := float64(0)

	for _, block := range bc.Pbc.GetBlocks() {
		for _, tx := range block.GetTxs() {
			if tx.GetRecipient() == pubkey {
				credit += tx.GetValue()
			} else if tx.GetSender() == pubkey {
				credit -= tx.GetValue()
			}
		}
	}

	return credit
}
