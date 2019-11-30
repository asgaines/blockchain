package nodes

import (
	"context"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/asgaines/blockchain/chain"
	"github.com/asgaines/blockchain/chain/mocks"
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

	type nodeSetup struct {
		chain             *chain.Chain
		targetDurPerBlock time.Duration
		recalcPeriod      int
		difficulty        float64
	}

	type mockCalls struct {
		mine struct {
			blocks []*chain.Block
		}
		setTarget struct {
			difficulties []float64
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
		nodeSetup nodeSetup
		mockCalls mockCalls
		expected  expected
	}{
		{
			name: "Mining a single block with exact desired duration has no effect on difficulty",
			nodeSetup: nodeSetup{
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
				difficulty:        100,
			},
			mockCalls: mockCalls{
				mine: struct {
					blocks []*chain.Block
				}{
					blocks: []*chain.Block{
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100,
							},
						},
					},
				},
				setTarget: struct {
					difficulties []float64
				}{
					difficulties: []float64{100},
				},
				clearTxs: struct {
					times int
				}{
					times: 1,
				},
			},
			expected: expected{
				difficulty: 100,
			},
		},
		{
			name: "Mining a single block with half desired duration doubles the difficulty",
			nodeSetup: nodeSetup{
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
				difficulty:        100,
			},
			mockCalls: mockCalls{
				mine: struct {
					blocks []*chain.Block
				}{
					blocks: []*chain.Block{
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 50,
							},
						},
					},
				},
				setTarget: struct {
					difficulties []float64
				}{
					difficulties: []float64{200},
				},
				clearTxs: struct {
					times int
				}{
					times: 1,
				},
			},
			expected: expected{
				difficulty: 200,
			},
		},
		{
			name: "Mining 3 blocks with a recalc period of 3 adjusts the difficulty by the average of all 3, taking slightly longer than desired",
			nodeSetup: nodeSetup{
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
				recalcPeriod:      3,
				difficulty:        100,
			},
			mockCalls: mockCalls{
				mine: struct {
					blocks []*chain.Block
				}{
					blocks: []*chain.Block{
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 199,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 307,
							},
						},
					},
				},
				setTarget: struct {
					difficulties []float64
				}{
					difficulties: []float64{97.71986970715871},
				},
				clearTxs: struct {
					times int
				}{
					times: 3,
				},
			},
			expected: expected{
				difficulty: 97.71986970715871,
			},
		},
		{
			name: "Mining 3 blocks with a recalc period of 3 adjusts the difficulty by the average of all 3, taking slightly less time than desired",
			nodeSetup: nodeSetup{
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
				recalcPeriod:      3,
				difficulty:        100,
			},
			mockCalls: mockCalls{
				mine: struct {
					blocks []*chain.Block
				}{
					blocks: []*chain.Block{
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 103,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 201,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 298,
							},
						},
					},
				},
				setTarget: struct {
					difficulties []float64
				}{
					difficulties: []float64{100.67114093993514},
				},
				clearTxs: struct {
					times int
				}{
					times: 3,
				},
			},
			expected: expected{
				difficulty: 100.67114093993514,
			},
		},
		{
			name: "Mining 3 blocks with a recalc period of 3 adjusts the difficulty by the average of all 3, taking slightly less time than desired, non-zero genesis block time",
			nodeSetup: nodeSetup{
				chain: &chain.Chain{
					Pbc: &blockchain.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Timestamp: &timestamp.Timestamp{
									Seconds: 10000,
								},
							},
						},
					},
				},
				targetDurPerBlock: 100 * time.Second,
				recalcPeriod:      3,
				difficulty:        100,
			},
			mockCalls: mockCalls{
				mine: struct {
					blocks []*chain.Block
				}{
					blocks: []*chain.Block{
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 10103,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 10201,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 10298,
							},
						},
					},
				},
				setTarget: struct {
					difficulties []float64
				}{
					difficulties: []float64{100.67114093993514},
				},
				clearTxs: struct {
					times int
				}{
					times: 3,
				},
			},
			expected: expected{
				difficulty: 100.67114093993514,
			},
		},
		{
			name: "Mining 3 blocks with a recalc period of 2 adjusts the difficulty by the average of only the first two, no second recalc triggered",
			nodeSetup: nodeSetup{
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
				recalcPeriod:      2,
				difficulty:        100,
			},
			mockCalls: mockCalls{
				mine: struct {
					blocks []*chain.Block
				}{
					blocks: []*chain.Block{
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 45,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 100,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 1000000,
							},
						},
					},
				},
				setTarget: struct {
					difficulties []float64
				}{
					difficulties: []float64{200},
				},
				clearTxs: struct {
					times int
				}{
					times: 3,
				},
			},
			expected: expected{
				difficulty: 200,
			},
		},
		{
			name: "Two recalc events triggered",
			nodeSetup: nodeSetup{
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
				recalcPeriod:      2,
				difficulty:        100,
			},
			mockCalls: mockCalls{
				mine: struct {
					blocks []*chain.Block
				}{
					blocks: []*chain.Block{
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 108,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 300,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 325,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 400,
							},
						},
					},
				},
				setTarget: struct {
					difficulties []float64
				}{
					difficulties: []float64{
						100 * (float64(2) / float64(3)),
						(100 * (float64(2) / float64(3))) * 2,
					},
				},
				clearTxs: struct {
					times int
				}{
					times: 4,
				},
			},
			expected: expected{
				difficulty: (100 * (float64(2) / float64(3))) * 2,
			},
		},
		{
			name: "Sub-second duration, half expected duration doubles the difficulty",
			nodeSetup: nodeSetup{
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
				targetDurPerBlock: 100 * time.Millisecond,
				recalcPeriod:      1,
				difficulty:        1000,
			},
			mockCalls: mockCalls{
				mine: struct {
					blocks []*chain.Block
				}{
					blocks: []*chain.Block{
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 0,
								Nanos:   50_000_000,
							},
						},
					},
				},
				setTarget: struct {
					difficulties []float64
				}{
					difficulties: []float64{
						2000,
					},
				},
				clearTxs: struct {
					times int
				}{
					times: 1,
				},
			},
			expected: expected{
				difficulty: 2000,
			},
		},
		{
			name: "5 difficulty adjustments for 5 solves; period of 1",
			nodeSetup: nodeSetup{
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
				targetDurPerBlock: 10 * time.Millisecond,
				recalcPeriod:      1,
				difficulty:        1000,
			},
			mockCalls: mockCalls{
				mine: struct {
					blocks []*chain.Block
				}{
					blocks: []*chain.Block{
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 0,
								Nanos:   20_000_000,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 0,
								Nanos:   30_000_000,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 0,
								Nanos:   45_000_000,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 0,
								Nanos:   55_000_000,
							},
						},
						&chain.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 0,
								Nanos:   60_000_000,
							},
						},
					},
				},
				setTarget: struct {
					difficulties []float64
				}{
					difficulties: []float64{
						500,
						500,
						333 + (float64(1) / float64(3)),
						333 + (float64(1) / float64(3)),
						666 + (float64(2) / float64(3)),
					},
				},
				clearTxs: struct {
					times int
				}{
					times: 5,
				},
			},
			expected: expected{
				difficulty: 666 + (float64(2) / float64(3)),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			mockMiner.EXPECT().Mine(ctx, gomock.Any()).
				Do(func(ctx context.Context, conveyor chan<- *chain.Block) {
					defer cancel()

					for _, block := range c.mockCalls.mine.blocks {
						conveyor <- block
					}

					close(conveyor)
				})
			for _, difficulty := range c.mockCalls.setTarget.difficulties {
				mockMiner.EXPECT().SetTarget(difficulty)
			}
			mockMiner.EXPECT().ClearTxs().Times(c.mockCalls.clearTxs.times)

			n := &node{
				chain:             c.nodeSetup.chain,
				targetDurPerBlock: c.nodeSetup.targetDurPerBlock,
				recalcPeriod:      c.nodeSetup.recalcPeriod,
				difficulty:        c.nodeSetup.difficulty,
				miner:             mockMiner,
			}

			n.mine(ctx)

			if n.difficulty != c.expected.difficulty {
				t.Errorf("expected %v got %v", c.expected.difficulty, n.difficulty)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockHasher := mocks.NewMockHasher(ctrl)

	type mockHashCall struct {
		in    *chain.Block
		out   []byte
		times int
	}

	cases := []struct {
		name          string
		chain         *chain.Chain
		mockHashCalls []mockHashCall
		want          bool
	}{
		{
			name: "Empty chain is not valid",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*blockchain.Block{},
				},
			},
			mockHashCalls: []mockHashCall{},
			want:          false,
		},
		{
			name: "Chain with valid hashing and prev hash reference is valid",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 646459200,
							},
							Hash: new(big.Int).SetUint64(1948111840464954436).Bytes(),
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 646469200,
							},
							Prevhash: new(big.Int).SetUint64(1948111840464954436).Bytes(),
							Nonce:    12345,
							Target:   new(big.Int).SetUint64(18446744073709551615).Bytes(),
							Pubkey:   "abc123",
							Hash:     new(big.Int).SetUint64(13857702854592346750).Bytes(),
						},
					},
				},
			},
			mockHashCalls: []mockHashCall{
				{
					in: &chain.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646459200,
						},
						Hash: new(big.Int).SetUint64(1948111840464954436).Bytes(),
					},
					out:   new(big.Int).SetUint64(1948111840464954436).Bytes(),
					times: 1,
				},
				{
					in: &chain.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646469200,
						},
						Prevhash: new(big.Int).SetUint64(1948111840464954436).Bytes(),
						Nonce:    12345,
						Target:   new(big.Int).SetUint64(18446744073709551615).Bytes(),
						Pubkey:   "abc123",
						Hash:     new(big.Int).SetUint64(13857702854592346750).Bytes(),
					},
					out:   new(big.Int).SetUint64(13857702854592346750).Bytes(),
					times: 1,
				},
			},
			want: true,
		},
		{
			name: "Chain with hash reported differently from actual hash result is not valid",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 646459200,
							},
							Hash: new(big.Int).SetUint64(1948111840464954436).Bytes(),
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 646469200,
							},
							Prevhash: new(big.Int).SetUint64(1948111840464954436).Bytes(),
							Nonce:    12345,
							Target:   new(big.Int).SetUint64(18446744073709551615).Bytes(),
							Pubkey:   "abc123",
							Hash:     new(big.Int).SetUint64(13857702854592346751).Bytes(),
						},
					},
				},
			},
			mockHashCalls: []mockHashCall{
				{
					in: &chain.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646459200,
						},
						Hash: new(big.Int).SetUint64(1948111840464954436).Bytes(),
					},
					out:   new(big.Int).SetUint64(1948111840464954436).Bytes(),
					times: 1,
				},
				{
					in: &chain.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646469200,
						},
						Prevhash: new(big.Int).SetUint64(1948111840464954436).Bytes(),
						Nonce:    12345,
						Target:   new(big.Int).SetUint64(18446744073709551615).Bytes(),
						Pubkey:   "abc123",
						Hash:     new(big.Int).SetUint64(13857702854592346751).Bytes(),
					},
					out:   new(big.Int).SetUint64(13857702854592346750).Bytes(),
					times: 1,
				},
			},
			want: false,
		},
		{
			name: "Chain with valid hash but missing (overshooting) the difficulty's target is not valid",
			chain: &chain.Chain{
				Pbc: &blockchain.Chain{
					Blocks: []*pb.Block{
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 646459200,
							},
							Hash: new(big.Int).SetUint64(1948111840464954436).Bytes(),
						},
						&pb.Block{
							Timestamp: &timestamp.Timestamp{
								Seconds: 646469200,
							},
							Prevhash: new(big.Int).SetUint64(1948111840464954436).Bytes(),
							Nonce:    123456789,
							Target:   new(big.Int).SetUint64(0).Bytes(), // Hardest possible target, requires full hash collision
							Pubkey:   "abc123",
							Hash:     new(big.Int).SetUint64(16295015879318905250).Bytes(),
						},
					},
				},
			},
			mockHashCalls: []mockHashCall{
				{
					in: &chain.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646459200,
						},
						Hash: new(big.Int).SetUint64(1948111840464954436).Bytes(),
					},
					out:   new(big.Int).SetUint64(1948111840464954436).Bytes(),
					times: 1,
				},
				{
					in: &chain.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646469200,
						},
						Prevhash: new(big.Int).SetUint64(1948111840464954436).Bytes(),
						Nonce:    123456789,
						Target:   new(big.Int).SetUint64(0).Bytes(), // Hardest possible target, requires full hash collision
						Pubkey:   "abc123",
						Hash:     new(big.Int).SetUint64(16295015879318905250).Bytes(),
					},
					out:   new(big.Int).SetUint64(16295015879318905250).Bytes(),
					times: 1,
				},
			},
			want: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, call := range c.mockHashCalls {
				mockHasher.EXPECT().Hash(call.in).Return(call.out).Times(call.times)
			}

			n := node{
				hasher: mockHasher,
			}

			got := n.IsValid(c.chain)

			if got != c.want {
				t.Errorf("want %v, got %v", c.want, got)
			}
		})
	}
}

func TestSetChain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMiner := mm.NewMockMiner(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)

	type nodeSetup struct {
		chain        *chain.Chain
		recalcPeriod int
	}

	type input struct {
		chain   *chain.Chain
		trusted bool
	}

	type mockHashCall struct {
		in    *chain.Block
		out   []byte
		times int
	}

	type mockMinerCalls struct {
		numClearTxs  int
		numSetTarget int
	}

	cases := []struct {
		name            string
		nodeSetup       nodeSetup
		input           input
		mockHashCalls   []mockHashCall
		mockMinerCalls  mockMinerCalls
		expectedReplace bool
	}{
		{
			name: "Valid chain of equal length does not replace node chain",
			nodeSetup: nodeSetup{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
						},
					},
				},
				recalcPeriod: 1,
			},
			input: input{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
						},
					},
				},
				trusted: false,
			},
			mockHashCalls: []mockHashCall{},
			mockMinerCalls: mockMinerCalls{
				numClearTxs:  0,
				numSetTarget: 0,
			},
			expectedReplace: false,
		},
		{
			name: "Trusted chain (also valid) of equal length does not replace node chain",
			nodeSetup: nodeSetup{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
						},
					},
				},
				recalcPeriod: 1,
			},
			input: input{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
						},
					},
				},
				trusted: true,
			},
			mockHashCalls: []mockHashCall{},
			mockMinerCalls: mockMinerCalls{
				numClearTxs:  0,
				numSetTarget: 0,
			},
			expectedReplace: false,
		},
		{
			name: "Valid chain of equal length replaces old chain",
			nodeSetup: nodeSetup{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
						},
					},
				},
				recalcPeriod: 1,
			},
			input: input{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
							&pb.Block{
								Hash:     []byte{2, 3, 4},
								Prevhash: []byte{1, 2, 3},
							},
						},
					},
				},
				trusted: false,
			},
			mockHashCalls: []mockHashCall{
				{
					in: &chain.Block{
						Hash: []byte{1, 2, 3},
					},
					out:   []byte{1, 2, 3},
					times: 1,
				},
				{
					in: &chain.Block{
						Hash:     []byte{2, 3, 4},
						Prevhash: []byte{1, 2, 3},
					},
					out:   []byte{2, 3, 4},
					times: 1,
				},
			},
			mockMinerCalls: mockMinerCalls{
				numClearTxs:  0,
				numSetTarget: 0,
			},
			expectedReplace: false,
		},
		{
			name: "Invalid chain hash (correct length) does not replace old chain",
			nodeSetup: nodeSetup{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
						},
					},
				},
				recalcPeriod: 1,
			},
			input: input{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
							&pb.Block{
								Hash:     []byte{9, 9, 9},
								Prevhash: []byte{1, 2, 3},
							},
						},
					},
				},
				trusted: false,
			},
			mockHashCalls: []mockHashCall{
				{
					in: &chain.Block{
						Hash: []byte{1, 2, 3},
					},
					out:   []byte{1, 2, 3},
					times: 1,
				},
				{
					in: &chain.Block{
						Hash:     []byte{9, 9, 9},
						Prevhash: []byte{1, 2, 3},
					},
					out:   []byte{2, 3, 4},
					times: 1,
				},
			},
			mockMinerCalls: mockMinerCalls{
				numClearTxs:  0,
				numSetTarget: 0,
			},
			expectedReplace: false,
		},
		{
			name: "Valid chain (chain +1 of existing length) replaces old chain",
			nodeSetup: nodeSetup{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
						},
					},
				},
				recalcPeriod: 1,
			},
			input: input{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
							&pb.Block{
								Hash:     []byte{2, 3, 4},
								Prevhash: []byte{1, 2, 3},
								Target:   []byte{2, 3, 4},
							},
						},
					},
				},
				trusted: false,
			},
			mockHashCalls: []mockHashCall{
				{
					in: &chain.Block{
						Hash: []byte{1, 2, 3},
					},
					out:   []byte{1, 2, 3},
					times: 1,
				},
				{
					in: &chain.Block{
						Hash:     []byte{2, 3, 4},
						Prevhash: []byte{1, 2, 3},
						Target:   []byte{2, 3, 4},
					},
					out:   []byte{2, 3, 4},
					times: 1,
				},
			},
			mockMinerCalls: mockMinerCalls{
				numClearTxs:  1,
				numSetTarget: 1,
			},
			expectedReplace: true,
		},
		{
			name: "Valid chain (chain +3 of existing length) replaces old chain",
			nodeSetup: nodeSetup{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
						},
					},
				},
				recalcPeriod: 1,
			},
			input: input{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
							&pb.Block{
								Hash:     []byte{2, 3, 4},
								Prevhash: []byte{1, 2, 3},
								Target:   []byte{2, 3, 4},
							},
							&pb.Block{
								Hash:     []byte{3, 4, 5},
								Prevhash: []byte{2, 3, 4},
								Target:   []byte{3, 4, 5},
							},
							&pb.Block{
								Hash:     []byte{4, 5, 6},
								Prevhash: []byte{3, 4, 5},
								Target:   []byte{4, 5, 6},
							},
						},
					},
				},
				trusted: false,
			},
			mockHashCalls: []mockHashCall{
				{
					in: &chain.Block{
						Hash:     []byte{4, 5, 6},
						Prevhash: []byte{3, 4, 5},
						Target:   []byte{4, 5, 6},
					},
					out:   []byte{4, 5, 6},
					times: 1,
				},
				{
					in: &chain.Block{
						Hash:     []byte{3, 4, 5},
						Prevhash: []byte{2, 3, 4},
						Target:   []byte{3, 4, 5},
					},
					out:   []byte{3, 4, 5},
					times: 1,
				},
				{
					in: &chain.Block{
						Hash:     []byte{3, 4, 5},
						Prevhash: []byte{2, 3, 4},
						Target:   []byte{3, 4, 5},
					},
					out:   []byte{3, 4, 5},
					times: 1,
				},
				{
					in: &chain.Block{
						Hash:     []byte{2, 3, 4},
						Prevhash: []byte{1, 2, 3},
						Target:   []byte{2, 3, 4},
					},
					out:   []byte{2, 3, 4},
					times: 1,
				},
				{
					in: &chain.Block{
						Hash:     []byte{2, 3, 4},
						Prevhash: []byte{1, 2, 3},
						Target:   []byte{2, 3, 4},
					},
					out:   []byte{2, 3, 4},
					times: 1,
				},
				{
					in: &chain.Block{
						Hash: []byte{1, 2, 3},
					},
					out:   []byte{1, 2, 3},
					times: 1,
				},
			},
			mockMinerCalls: mockMinerCalls{
				numClearTxs:  1,
				numSetTarget: 1,
			},
			expectedReplace: true,
		},
		{
			name: "Valid chain (chain +1 of existing length) replaces old chain, does not retrigger difficulty recalc when period not matched",
			nodeSetup: nodeSetup{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
						},
					},
				},
				recalcPeriod: 2,
			},
			input: input{
				chain: &chain.Chain{
					Pbc: &pb.Chain{
						Blocks: []*pb.Block{
							&pb.Block{
								Hash: []byte{1, 2, 3},
							},
							&pb.Block{
								Hash:     []byte{2, 3, 4},
								Prevhash: []byte{1, 2, 3},
								Target:   []byte{2, 3, 4},
							},
						},
					},
				},
				trusted: false,
			},
			mockHashCalls: []mockHashCall{
				{
					in: &chain.Block{
						Hash: []byte{1, 2, 3},
					},
					out:   []byte{1, 2, 3},
					times: 1,
				},
				{
					in: &chain.Block{
						Hash:     []byte{2, 3, 4},
						Prevhash: []byte{1, 2, 3},
						Target:   []byte{2, 3, 4},
					},
					out:   []byte{2, 3, 4},
					times: 1,
				},
			},
			mockMinerCalls: mockMinerCalls{
				numClearTxs:  1,
				numSetTarget: 0,
			},
			expectedReplace: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, call := range c.mockHashCalls {
				mockHasher.EXPECT().Hash(call.in).Return(call.out).Times(call.times)
			}

			mockMiner.EXPECT().SetTarget(gomock.Any()).Times(c.mockMinerCalls.numSetTarget)
			mockMiner.EXPECT().ClearTxs().Times(c.mockMinerCalls.numClearTxs)

			n := node{
				chain:        c.nodeSetup.chain,
				recalcPeriod: c.nodeSetup.recalcPeriod,
				miner:        mockMiner,
				hasher:       mockHasher,
			}

			got := n.setChain(c.input.chain, c.input.trusted)

			if got != c.expectedReplace {
				t.Errorf("expected replacement: %v, got %v", c.expectedReplace, got)
			}

			if c.expectedReplace {
				if !reflect.DeepEqual(n.chain, c.input.chain) {
					t.Errorf("expected chain: %v, got %v", c.input.chain, n.chain)
				}
			} else {
				if !reflect.DeepEqual(n.chain, c.nodeSetup.chain) {
					t.Errorf("expected chain: %v, got %v", c.nodeSetup.chain, n.chain)
				}
			}
		})
	}
}

func TestCalcDifficulty(t *testing.T) {
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
			expected:       682 + float64(2)/float64(3),
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
		{
			name: "Actual duration very close (slightly longer) to desired adjusts slightly, highlighting math accuracy",
			node: node{
				targetDurPerBlock: 10 * time.Minute,
			},
			actualDur:      10*time.Minute + 4*time.Second + 563*time.Millisecond,
			currDifficulty: 623503.2744,
			expected:       618797.3207755022,
		},
		{
			name: "Actual duration very close (slightly shorter) to desired adjusts slightly, highlighting math accuracy",
			node: node{
				targetDurPerBlock: 10 * time.Millisecond,
			},
			actualDur:      9*time.Millisecond + 981_613*time.Nanosecond,
			currDifficulty: 1674902436.200243,
			expected:       1677987752.280361,
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

func TestConfine(t *testing.T) {
	cases := []struct {
		name     string
		in       float64
		expected float64
	}{
		{
			name:     "An adjustment of 1 (no change) is allowed",
			in:       1,
			expected: 1,
		},
		{
			name:     "A halving adjustment is allowed",
			in:       float64(1) / float64(2),
			expected: float64(1) / float64(2),
		},
		{
			name:     "A doubling adjustment is allowed",
			in:       2,
			expected: 2,
		},
		{
			name:     "An adjustment of multiplying by 5 is confined to 4",
			in:       5,
			expected: 4,
		},
		{
			name:     "An adjustment of division by 5 is confined to 4",
			in:       float64(1) / float64(5),
			expected: float64(1) / float64(4),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n := node{}
			got := n.confine(c.in)
			if got != c.expected {
				t.Errorf("expected %v, got %v", c.expected, got)
			}
		})
	}
}
