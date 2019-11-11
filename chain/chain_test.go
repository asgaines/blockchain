package chain

import (
	"testing"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func TestIsSolid(t *testing.T) {
	cases := []struct {
		message string
		chain   *Chain
		want    bool
	}{
		{
			message: "Chain with only genesis block is always solid",
			chain:   NewChain(),
			want:    true,
		},
		{
			message: "Empty chain is not solid",
			chain:   &Chain{},
			want:    false,
		},
		{
			message: "Chain with valid hashing and prev hash reference is solid",
			chain: &Chain{
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
			want: true,
		},
		{
			message: "Chain with valid prev hash reference but misreported hash is not solid",
			chain: &Chain{
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
						Hash:     13857702854592346751, // should be 13857702854592346750
					},
				},
			},
			want: false,
		},
		{
			message: "Chain with valid hash but missing difficulty's target is not solid",
			chain: &Chain{
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
			want: false,
		},
	}

	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			got := c.chain.IsSolid()

			if got != c.want {
				t.Errorf("want %v, got %v", c.want, got)
			}
		})
	}
}
