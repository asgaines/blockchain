package chain

import (
	"math/big"
	"reflect"
	"testing"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

func TestLastLink(t *testing.T) {
	cases := []struct {
		name     string
		chain    *Chain
		expected *Block
	}{
		{
			name: "A chain with only a single block has it returned",
			chain: &Chain{
				Pbc: &pb.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Hash: new(big.Int).SetInt64(12345).Bytes(),
						},
					},
				},
			},
			expected: &Block{
				Hash: new(big.Int).SetInt64(12345).Bytes(),
			},
		},
		{
			name: "A chain with two blocks has the last returned",
			chain: &Chain{
				Pbc: &pb.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Hash: new(big.Int).SetInt64(12345).Bytes(),
						},
						&pb.Block{
							Hash: new(big.Int).SetInt64(23456).Bytes(),
						},
					},
				},
			},
			expected: &Block{
				Hash: new(big.Int).SetInt64(23456).Bytes(),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.chain.LastLink()

			if !reflect.DeepEqual(got, c.expected) {
				t.Errorf("expected %+v, got %+v", c.expected, got)
			}
		})
	}
}

func TestLength(t *testing.T) {
	cases := []struct {
		name     string
		chain    *Chain
		expected int
	}{
		{
			name: "An empty chain has length 0",
			chain: &Chain{
				Pbc: &pb.Chain{
					Blocks: []*pb.Block{},
				},
			},
			expected: 0,
		},
		{
			name: "A chain with only a single block has length 1",
			chain: &Chain{
				Pbc: &pb.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Hash: new(big.Int).SetInt64(12345).Bytes(),
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "A chain with two blocks has length 2",
			chain: &Chain{
				Pbc: &pb.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Hash: new(big.Int).SetInt64(12345).Bytes(),
						},
						&pb.Block{
							Hash: new(big.Int).SetInt64(23456).Bytes(),
						},
					},
				},
			},
			expected: 2,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.chain.Length()

			if !reflect.DeepEqual(got, c.expected) {
				t.Errorf("expected %+v, got %+v", c.expected, got)
			}
		})
	}
}