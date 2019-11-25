package chain_test

import (
	"testing"

	"github.com/asgaines/blockchain/chain"
	"github.com/asgaines/blockchain/chain/mocks"
	"github.com/asgaines/blockchain/protogo/blockchain"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func TestIsSolid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockHasher := mocks.NewMockHasher(ctrl)

	type mockHashCall struct {
		in    *chain.Block
		out   uint64
		times int
	}

	cases := []struct {
		name          string
		pchain        *blockchain.Chain
		mockHashCalls []mockHashCall
		want          bool
	}{
		{
			name: "Empty chain is not valid",
			pchain: &blockchain.Chain{
				Blocks: []*blockchain.Block{},
			},
			mockHashCalls: []mockHashCall{},
			want:          false,
		},
		{
			name: "Chain with valid hashing and prev hash reference is valid",
			pchain: &blockchain.Chain{
				Blocks: []*pb.Block{
					&pb.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646459200,
						},
						Hash: 1948111840464954436,
					},
					&pb.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646469200,
						},
						Prevhash: 1948111840464954436,
						Nonce:    12345,
						Target:   18446744073709551615,
						Pubkey:   "abc123",
						Hash:     13857702854592346750,
					},
				},
			},
			mockHashCalls: []mockHashCall{
				{
					in: &chain.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646459200,
						},
						Hash: 1948111840464954436,
					},
					out:   1948111840464954436,
					times: 1,
				},
				{
					in: &chain.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646469200,
						},
						Prevhash: 1948111840464954436,
						Nonce:    12345,
						Target:   18446744073709551615,
						Pubkey:   "abc123",
						Hash:     13857702854592346750,
					},
					out:   13857702854592346750,
					times: 1,
				},
			},
			want: true,
		},
		{
			name: "Chain with hash reported differently from actual hash result is not valid",
			pchain: &blockchain.Chain{
				Blocks: []*pb.Block{
					&pb.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646459200,
						},
						Hash: 1948111840464954436,
					},
					&pb.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646469200,
						},
						Prevhash: 1948111840464954436,
						Nonce:    12345,
						Target:   18446744073709551615,
						Pubkey:   "abc123",
						Hash:     13857702854592346751,
					},
				},
			},
			mockHashCalls: []mockHashCall{
				{
					in: &chain.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646459200,
						},
						Hash: 1948111840464954436,
					},
					out:   1948111840464954436,
					times: 1,
				},
				{
					in: &chain.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646469200,
						},
						Prevhash: 1948111840464954436,
						Nonce:    12345,
						Target:   18446744073709551615,
						Pubkey:   "abc123",
						Hash:     13857702854592346751,
					},
					out:   13857702854592346750,
					times: 1,
				},
			},
			want: false,
		},
		{
			name: "Chain with valid hash but missing (overshooting) the difficulty's target is not valid",
			pchain: &blockchain.Chain{
				Blocks: []*pb.Block{
					&pb.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646459200,
						},
						Hash: 1948111840464954436,
					},
					&pb.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646469200,
						},
						Prevhash: 1948111840464954436,
						Nonce:    123456789,
						Target:   0, // Hardest possible target, requires full hash collision
						Pubkey:   "abc123",
						Hash:     16295015879318905250,
					},
				},
			},
			mockHashCalls: []mockHashCall{
				{
					in: &chain.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646459200,
						},
						Hash: 1948111840464954436,
					},
					out:   1948111840464954436,
					times: 1,
				},
				{
					in: &chain.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 646469200,
						},
						Prevhash: 1948111840464954436,
						Nonce:    123456789,
						Target:   0, // Hardest possible target, requires full hash collision
						Pubkey:   "abc123",
						Hash:     16295015879318905250,
					},
					out:   16295015879318905250,
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

			ch := &chain.Chain{
				Pbc:    c.pchain,
				Hasher: mockHasher,
			}

			got := ch.IsSolid()

			if got != c.want {
				t.Errorf("want %v, got %v", c.want, got)
			}
		})
	}
}
