package nodes

import (
	"context"
	"testing"
	"time"

	"github.com/asgaines/blockchain/chain"
	mm "github.com/asgaines/blockchain/mining/mocks"
	"github.com/asgaines/blockchain/protogo/blockchain"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func TestMine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMiner := mm.NewMockMiner(ctrl)

	type mockCalls struct {
		mine struct {
			blocks []*chain.Block
			times  int
		}
		setTarget struct {
			difficulty float64
			times      int
		}
		clearTxs struct {
			times int
		}
	}

	type expected struct {
		difficulty float64
	}

	cases := []struct {
		name      string
		node      *node
		mockCalls mockCalls
		expected  expected
	}{
		{
			name: "Mining a single block with exact desired duration has no effect on difficulty",
			node: &node{
				chain: &chain.Chain{
					Pbc: &blockchain.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Timestamp: &timestamp.Timestamp{
									Seconds: 0,
								},
							},
						},
					},
				},
				targetDurPerBlock: 100 * time.Second,
				recalcPeriod:      1,
				difficulty:        1,
			},
			mockCalls: mockCalls{
				mine: struct {
					blocks []*chain.Block
					times  int
				}{
					blocks: []*chain.Block{
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100,
							},
						},
					},
					times: 1,
				},
				setTarget: struct {
					difficulty float64
					times      int
				}{
					difficulty: 1,
					times:      1,
				},
				clearTxs: struct {
					times int
				}{
					times: 1,
				},
			},
			expected: expected{
				difficulty: 1,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			mockMiner.EXPECT().Mine(gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, conveyor chan<- *chain.Block) {
					defer cancel()

					for _, block := range c.mockCalls.mine.blocks {
						conveyor <- block
					}

					close(conveyor)
				}).Times(c.mockCalls.mine.times)
			mockMiner.EXPECT().SetTarget(c.mockCalls.setTarget.difficulty).Times(c.mockCalls.setTarget.times)
			mockMiner.EXPECT().ClearTxs().Times(c.mockCalls.clearTxs.times)

			c.node.miner = mockMiner

			c.node.mine(ctx)

			if c.node.difficulty != c.expected.difficulty {
				t.Errorf("expected %v got %v", c.expected.difficulty, c.node.difficulty)
			}
		})
	}
}

func TestGetDifficulty(t *testing.T) {
	cases := []struct {
		name           string
		node           node
		actualDur      time.Duration
		currDifficulty float64
		expected       float64
	}{
		{
			name: "An exact match between actual and desired duration returns the same difficulty",
			node: node{
				targetDurPerBlock: 10 * time.Minute,
			},
			actualDur:      10 * time.Minute,
			currDifficulty: 1024,
			expected:       1024,
		},
		{
			name: "An actual duration half of expected returns a difficulty twice of the current value",
			node: node{
				targetDurPerBlock: 10 * time.Minute,
			},
			actualDur:      5 * time.Minute,
			currDifficulty: 1024,
			expected:       2048,
		},
		{
			name: "An actual duration twice of expected returns a difficulty half of the current value",
			node: node{
				targetDurPerBlock: 10 * time.Minute,
			},
			actualDur:      20 * time.Minute,
			currDifficulty: 1024,
			expected:       512,
		},
		{
			name: "An actual duration 1.5 times of expected returns a difficulty quotient of 1.5 of the current value",
			node: node{
				targetDurPerBlock: 10 * time.Minute,
			},
			actualDur:      15 * time.Minute,
			currDifficulty: 1024,
			expected:       1024 / 1.5,
		},
		{
			name: "An actual duration 10 times of expected returns a difficulty confined to 1/4 the previous amount, even though the calculation would be 1/10",
			node: node{
				targetDurPerBlock: 10 * time.Minute,
			},
			actualDur:      100 * time.Minute,
			currDifficulty: 1024,
			expected:       1024 / 4,
		},
		{
			name: "An actual duration 1/10 of expected returns a difficulty confined to 4 times the previous amount, even though the calculation would be x10",
			node: node{
				targetDurPerBlock: 10 * time.Minute,
			},
			actualDur:      1 * time.Minute,
			currDifficulty: 1024,
			expected:       1024 * 4,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.node.calcDifficulty(c.actualDur, c.currDifficulty)

			if got != c.expected {
				t.Errorf("expected %v, got %v", c.expected, got)
			}
		})
	}
}

func TestGetRecalcRangeDur(t *testing.T) {
	type expected struct {
		dur    time.Duration
		hasErr bool
	}

	cases := []struct {
		name         string
		chain        *chain.Chain
		recalcPeriod int
		expected     expected
	}{
		{
			name: "A recalc period of 1 is the duration between the two links",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000010,
							},
						},
					},
				},
			},
			recalcPeriod: 1,
			expected: expected{
				dur:    10 * time.Second,
				hasErr: false,
			},
		},
		{
			name: "A recalc period of greater than the length of the chain yields an error",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000010,
							},
						},
					},
				},
			},
			recalcPeriod: 3,
			expected: expected{
				dur:    0 * time.Second,
				hasErr: true,
			},
		},
		{
			name: "A recalc period > 1 calculates the difference between the correct blocks",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000005,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000011,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000018,
							},
						},
					},
				},
			},
			recalcPeriod: 2,
			expected: expected{
				dur:    13 * time.Second,
				hasErr: false,
			},
		},
	}

	n := node{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := n.getRecalcRangeDur(c.chain, c.recalcPeriod)
			if (err != nil) != c.expected.hasErr {
				t.Errorf("expected error: %v, got %v", c.expected.hasErr, err)
			}

			if got != c.expected.dur {
				t.Errorf("expected %v, got %v", c.expected.dur, got)
			}
		})
	}
}

func TestGetLastBlockDur(t *testing.T) {
	type expected struct {
		dur    time.Duration
		hasErr bool
	}

	cases := []struct {
		name         string
		chain        *chain.Chain
		recalcPeriod int
		expected     expected
	}{
		{
			name: "The last solve duration is the difference between the two blocks for a chain of length 2",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000010,
							},
						},
					},
				},
			},
			expected: expected{
				dur:    10 * time.Second,
				hasErr: false,
			},
		},
		{
			name: "The last solve duration correct for a chain of length > 2",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000010,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000021,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000033,
							},
						},
					},
				},
			},
			expected: expected{
				dur:    12 * time.Second,
				hasErr: false,
			},
		},
		{
			name: "Getting solve duration when only genesis block is present is an error",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
							},
						},
					},
				},
			},
			expected: expected{
				dur:    0 * time.Second,
				hasErr: true,
			},
		},
	}

	n := node{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := n.getLastBlockDur(c.chain)
			if (err != nil) != c.expected.hasErr {
				t.Errorf("expected error: %v, got %v", c.expected.hasErr, err)
			}

			if got != c.expected.dur {
				t.Errorf("expected %v, got %v", c.expected.dur, got)
			}
		})
	}
}

func TestGetRangeAvgBlockDur(t *testing.T) {
	type expected struct {
		dur    time.Duration
		hasErr bool
	}

	cases := []struct {
		name         string
		chain        *chain.Chain
		recalcPeriod int
		expected     expected
	}{
		{
			name: "A recalc period of 1 is the duration between the two links",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000010,
							},
						},
					},
				},
			},
			recalcPeriod: 1,
			expected: expected{
				dur:    10 * time.Second,
				hasErr: false,
			},
		},
		{
			name: "A recalc period of 2 averages the 2 solve durations",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000010,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000030,
							},
						},
					},
				},
			},
			recalcPeriod: 2,
			expected: expected{
				dur:    15 * time.Second,
				hasErr: false,
			},
		},
		{
			name: "A recalc period of 3 with a wide range of times yields the right average",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000012,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000013,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000054,
							},
						},
					},
				},
			},
			recalcPeriod: 3,
			expected: expected{
				dur:    18 * time.Second,
				hasErr: false,
			},
		},
		{
			name: "A recalc period of 4 with an expected output of a fractional second is correct",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000012,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000013,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000054,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000057,
							},
						},
					},
				},
			},
			recalcPeriod: 4,
			expected: expected{
				dur:    14*time.Second + 250*time.Millisecond,
				hasErr: false,
			},
		},
		{
			name: "An average solve duration for a single sub-second block is correct",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
								Nanos:   0,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
								Nanos:   500000000,
							},
						},
					},
				},
			},
			recalcPeriod: 1,
			expected: expected{
				dur:    500 * time.Millisecond,
				hasErr: false,
			},
		},
		{
			name: "An average solve duration of more than one sub-second block is correct",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
								Nanos:   0,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
								Nanos:   400000000,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
								Nanos:   600000000,
							},
						},
					},
				},
			},
			recalcPeriod: 2,
			expected: expected{
				dur:    300 * time.Millisecond,
				hasErr: false,
			},
		},
		{
			name: "An average solve duration of more than one sub-second block is correct with an odd-number value",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
								Nanos:   0,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
								Nanos:   348000000,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
								Nanos:   730000000,
							},
						},
					},
				},
			},
			recalcPeriod: 2,
			expected: expected{
				dur:    365 * time.Millisecond,
				hasErr: false,
			},
		},
		{
			name: "A chain with length greater than the recalc period has the correct durations used in calculation",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000000,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000010,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000021,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000033,
							},
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100000046,
							},
						},
					},
				},
			},
			recalcPeriod: 2,
			expected: expected{
				dur:    12*time.Second + 500*time.Millisecond,
				hasErr: false,
			},
		},
	}

	n := node{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := n.getRangeAvgBlockDur(c.chain, c.recalcPeriod)
			if (err != nil) != c.expected.hasErr {
				t.Errorf("expected error: %v, got %v", c.expected.hasErr, err)
			}

			if got != c.expected.dur {
				t.Errorf("expected %v, got %v", c.expected.dur, got)
			}
		})
	}
}
