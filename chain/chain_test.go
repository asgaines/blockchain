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
						{
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
						{
							Hash: new(big.Int).SetInt64(12345).Bytes(),
						},
						{
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
						{
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
						{
							Hash: new(big.Int).SetInt64(12345).Bytes(),
						},
						{
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

func TestGetCreditFor(t *testing.T) {
	cases := []struct {
		name     string
		chain    *Chain
		pubkey   string
		expected float64
	}{
		{
			name: "An empty chain will have no credit",
			chain: &Chain{
				Pbc: &pb.Chain{
					Blocks: []*pb.Block{},
				},
			},
			pubkey:   "GothicCastle",
			expected: 0,
		},
		{
			name: "A single tx with the matching pubkey returns the value for that tx",
			chain: &Chain{
				Pbc: &pb.Chain{
					Blocks: []*pb.Block{
						{
							Txs: []*pb.Tx{
								{
									Recipient: "SteveHolt",
									Value:     5.5,
								},
							},
						},
					},
				},
			},
			pubkey:   "SteveHolt",
			expected: 5.5,
		},
		{
			name: "A single tx with different pubkey does not credit the key",
			chain: &Chain{
				Pbc: &pb.Chain{
					Blocks: []*pb.Block{
						{
							Txs: []*pb.Tx{
								{
									Recipient: "GeorgeMichael",
									Value:     12.5,
								},
							},
						},
					},
				},
			},
			pubkey:   "GeorgeSenior",
			expected: 0,
		},
		{
			name: "Two matching transactions in same block are added together for matching key",
			chain: &Chain{
				Pbc: &pb.Chain{
					Blocks: []*pb.Block{
						{
							Txs: []*pb.Tx{
								{
									Recipient: "LucilleBluth",
									Value:     12.5,
								},
								{
									Recipient: "LucilleBluth",
									Value:     100,
								},
							},
						},
					},
				},
			},
			pubkey:   "LucilleBluth",
			expected: 112.5,
		},
		{
			name: "Credits and debits are balanced against one another for the final difference",
			chain: &Chain{
				Pbc: &pb.Chain{
					Blocks: []*pb.Block{
						{
							Txs: []*pb.Tx{
								{
									Recipient: "Rita",
									Value:     70,
								},
							},
						},
						{
							Txs: []*pb.Tx{
								{
									Sender: "Rita",
									Value:  50,
								},
							},
						},
					},
				},
			},
			pubkey:   "Rita",
			expected: 20,
		},
		{
			name: "Two matching transactions in 2 different blocks are added together for matching key",
			chain: &Chain{
				Pbc: &pb.Chain{
					Blocks: []*pb.Block{
						{
							Txs: []*pb.Tx{
								{
									Recipient: "MrF",
									Value:     2.5,
								},
							},
						},
						{
							Txs: []*pb.Tx{
								{
									Recipient: "MrF",
									Value:     5,
								},
							},
						},
					},
				},
			},
			pubkey:   "MrF",
			expected: 7.5,
		},
		{
			name: "Many blocks with many transactions only return the accumulation of the matching key",
			chain: &Chain{
				Pbc: &pb.Chain{
					Blocks: []*pb.Block{
						{
							Txs: []*pb.Tx{
								{
									Recipient: "LindsayBluth",
									Value:     100,
								},
								{
									Recipient: "DirtyEarsBill",
									Value:     8,
								},
							},
						},
						{
							Txs: []*pb.Tx{
								{
									Recipient: "LindsayBluth",
									Value:     5,
								},
							},
						},
						{
							Txs: []*pb.Tx{
								{
									Recipient: "Gob",
									Value:     35,
								},
							},
						},
						{
							Txs: []*pb.Tx{
								{
									Recipient: "Hermano",
									Value:     7,
								},
								{
									Sender: "LindsayBluth",
									Value:  35,
								},
							},
						},
					},
				},
			},
			pubkey:   "LindsayBluth",
			expected: 70,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.chain.GetCreditFor(c.pubkey)

			if !reflect.DeepEqual(got, c.expected) {
				t.Errorf("expected %v, got %v", c.expected, got)
			}
		})
	}
}
