package mining

import (
	"context"
	"testing"
	"time"

	"github.com/asgaines/blockchain/chain"
	"github.com/asgaines/blockchain/chain/mocks"
	"github.com/golang/mock/gomock"
)

func TestMine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockHasher := mocks.NewMockHasher(ctrl)

	type minerSetup struct {
		prevBlock *chain.Block
		target    float64
	}

	type mockHashCall struct {
		out   uint64
		times int
	}

	type expected struct {
		numSolves int
	}

	cases := []struct {
		name          string
		minerSetup    minerSetup
		mockHashCalls []mockHashCall
		expected      expected
	}{
		{
			name: "Hashing to the exact target is a successful solve, first try",
			minerSetup: minerSetup{
				prevBlock: &chain.Block{},
				target:    1000,
			},
			mockHashCalls: []mockHashCall{
				{
					out:   1000,
					times: 1,
				},
			},
			expected: expected{
				numSolves: 1,
			},
		},
		{
			name: "Hashing to the exact target is a successful solve, third try",
			minerSetup: minerSetup{
				prevBlock: &chain.Block{},
				target:    1000,
			},
			mockHashCalls: []mockHashCall{
				{
					out:   48067290,
					times: 1,
				},
				{
					out:   712401678015,
					times: 1,
				},
				{
					out:   1000,
					times: 1,
				},
			},
			expected: expected{
				numSolves: 1,
			},
		},
		{
			name: "Lucky first hash, no luck afterwards",
			minerSetup: minerSetup{
				prevBlock: &chain.Block{},
				target:    1000,
			},
			mockHashCalls: []mockHashCall{
				{
					out:   999,
					times: 1,
				},
				{
					out:   712401678015,
					times: 1,
				},
				{
					out:   1001,
					times: 1,
				},
			},
			expected: expected{
				numSolves: 1,
			},
		},
		{
			name: "5 unsuccessful hashes yield no solves",
			minerSetup: minerSetup{
				prevBlock: &chain.Block{},
				target:    1000,
			},
			mockHashCalls: []mockHashCall{
				{
					out:   48067290,
					times: 1,
				},
				{
					out:   712401678015,
					times: 1,
				},
				{
					out:   1001,
					times: 1,
				},
				{
					out:   1002,
					times: 1,
				},
				{
					out:   1003,
					times: 1,
				},
			},
			expected: expected{
				numSolves: 0,
			},
		},
		{
			name: "5 successful hashes yields 5 solves",
			minerSetup: minerSetup{
				prevBlock: &chain.Block{},
				target:    1000,
			},
			mockHashCalls: []mockHashCall{
				{
					out:   1000,
					times: 1,
				},
				{
					out:   999,
					times: 1,
				},
				{
					out:   0,
					times: 1,
				},
				{
					out:   5,
					times: 1,
				},
				{
					out:   25,
					times: 1,
				},
			},
			expected: expected{
				numSolves: 5,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			conveyor := make(chan *chain.Block)

			for n, call := range c.mockHashCalls {
				func(lastCall bool, out uint64) {
					mockHasher.EXPECT().Hash(gomock.Any()).
						DoAndReturn(func(block *chain.Block) uint64 {
							if lastCall {
								cancel()
							}

							return out
						}).Times(call.times)
				}(n == len(c.mockHashCalls)-1, call.out)
			}

			m := &miner{
				prevBlock: c.minerSetup.prevBlock,
				target:    c.minerSetup.target,
				hasher:    mockHasher,
			}

			go m.Mine(ctx, conveyor)

			solves := []*chain.Block{}
			for b := range conveyor {
				solves = append(solves, b)
			}

			if len(solves) != c.expected.numSolves {
				t.Errorf("expected %d solve(s), got %d", c.expected.numSolves, len(solves))
			}
		})
	}
}

func TestSetTarget(t *testing.T) {
	cases := []struct {
		name       string
		difficulty float64
		expected   float64
	}{
		{
			name:       "Target is maximum (easiest) if difficulty is minimum",
			difficulty: 1,
			expected:   MaxTarget,
		},
		{
			name:       "Target is maximum (easiest) if difficulty is less than 1: edge case, since minimum difficulty is 1",
			difficulty: 0.1,
			expected:   MaxTarget,
		},
		{
			name:       "Target is 1/100th of total range for a difficulty of 100",
			difficulty: 100,
			expected:   MaxTarget / 100,
		},
	}

	m := miner{}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m.SetTarget(c.difficulty)

			if m.target != c.expected {
				t.Errorf("expected %v, got %v", c.expected, m.target)
			}
		})
	}
}
