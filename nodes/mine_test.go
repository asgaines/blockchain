package nodes

import (
	"testing"
	"time"

	"github.com/asgaines/blockchain/chain"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func TestGetRecalcRangeDur(t *testing.T) {
	type expected struct {
		dur time.Duration
		err bool
	}

	cases := []struct {
		message      string
		chain        *chain.Chain
		recalcPeriod int
		expected     expected
	}{
		{
			message: "A recalc period of 1 is the duration between the two links",
			chain: &chain.Chain{
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
			recalcPeriod: 1,
			expected: expected{
				dur: 10 * time.Second,
				err: false,
			},
		},
		{
			message: "A recalc period of greater than the length of the chain yields an error",
			chain: &chain.Chain{
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
			recalcPeriod: 3,
			expected: expected{
				dur: 0 * time.Second,
				err: true,
			},
		},
		{
			message: "A recalc period > 1 calculates the difference between the correct blocks",
			chain: &chain.Chain{
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
			recalcPeriod: 2,
			expected: expected{
				dur: 13 * time.Second,
				err: false,
			},
		},
	}

	n := node{}
	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			got, err := n.getRecalcRangeDur(c.chain, c.recalcPeriod)
			if (err != nil) != c.expected.err {
				t.Errorf("expected error: %v, got %v", c.expected.err, err)
			}

			if got != c.expected.dur {
				t.Errorf("expected %v, got %v", c.expected.dur, got)
			}
		})
	}
}

func TestGetLastBlockDur(t *testing.T) {
	type expected struct {
		dur time.Duration
		err bool
	}

	cases := []struct {
		message      string
		chain        *chain.Chain
		recalcPeriod int
		expected     expected
	}{
		{
			message: "The last solve duration is the difference between the two blocks for a chain of length 2",
			chain: &chain.Chain{
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
			expected: expected{
				dur: 10 * time.Second,
				err: false,
			},
		},
		{
			message: "The last solve duration correct for a chain of length > 2",
			chain: &chain.Chain{
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
			expected: expected{
				dur: 12 * time.Second,
				err: false,
			},
		},
		{
			message: "Getting solve duration when only genesis block is present is an error",
			chain: &chain.Chain{
				Blocks: []*pb.Block{
					&pb.Block{
						Timestamp: &timestamp.Timestamp{
							Seconds: 100000000,
						},
					},
				},
			},
			expected: expected{
				dur: 0 * time.Second,
				err: true,
			},
		},
	}

	n := node{}
	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			got, err := n.getLastBlockDur(c.chain)
			if (err != nil) != c.expected.err {
				t.Errorf("expected error: %v, got %v", c.expected.err, err)
			}

			if got != c.expected.dur {
				t.Errorf("expected %v, got %v", c.expected.dur, got)
			}
		})
	}
}

func TestGetRangeAvgBlockDur(t *testing.T) {
	type expected struct {
		dur time.Duration
		err bool
	}

	cases := []struct {
		message      string
		chain        *chain.Chain
		recalcPeriod int
		expected     expected
	}{
		{
			message: "A recalc period of 1 is the duration between the two links",
			chain: &chain.Chain{
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
			recalcPeriod: 1,
			expected: expected{
				dur: 10 * time.Second,
				err: false,
			},
		},
		{
			message: "A recalc period of 2 averages the 2 solve durations",
			chain: &chain.Chain{
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
			recalcPeriod: 2,
			expected: expected{
				dur: 15 * time.Second,
				err: false,
			},
		},
		{
			message: "A recalc period of 3 with a wide range of times yields the right average",
			chain: &chain.Chain{
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
			recalcPeriod: 3,
			expected: expected{
				dur: 18 * time.Second,
				err: false,
			},
		},
		{
			message: "A recalc period of 4 with an expected output of a fractional second is correct",
			chain: &chain.Chain{
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
			recalcPeriod: 4,
			expected: expected{
				dur: 14*time.Second + 250*time.Millisecond,
				err: false,
			},
		},
		{
			message: "An average solve duration for a single sub-second block is correct",
			chain: &chain.Chain{
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
			recalcPeriod: 1,
			expected: expected{
				dur: 500 * time.Millisecond,
				err: false,
			},
		},
		{
			message: "An average solve duration of more than one sub-second block is correct",
			chain: &chain.Chain{
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
			recalcPeriod: 2,
			expected: expected{
				dur: 300 * time.Millisecond,
				err: false,
			},
		},
		{
			message: "An average solve duration of more than one sub-second block is correct with an odd-number value",
			chain: &chain.Chain{
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
			recalcPeriod: 2,
			expected: expected{
				dur: 365 * time.Millisecond,
				err: false,
			},
		},
	}

	n := node{}
	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			got, err := n.getRangeAvgBlockDur(c.chain, c.recalcPeriod)
			if (err != nil) != c.expected.err {
				t.Errorf("expected error: %v, got %v", c.expected.err, err)
			}

			if got != c.expected.dur {
				t.Errorf("expected %v, got %v", c.expected.dur, got)
			}
		})
	}
}
