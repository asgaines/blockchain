package mining

import (
	"bytes"
	"context"
	"math/big"
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
		target    []byte
	}

	type mockHashCall struct {
		out   []byte
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
				target:    []byte{123, 234},
			},
			mockHashCalls: []mockHashCall{
				{
					out:   []byte{123, 234},
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
				target:    []byte{123, 234},
			},
			mockHashCalls: []mockHashCall{
				{
					out:   []byte{123, 234, 1},
					times: 1,
				},
				{
					out:   []byte{123, 234, 123},
					times: 1,
				},
				{
					out:   []byte{123, 234},
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
				target:    []byte{123, 234},
			},
			mockHashCalls: []mockHashCall{
				{
					out:   []byte{123, 233},
					times: 1,
				},
				{
					out:   []byte{51, 36, 201, 123, 233},
					times: 1,
				},
				{
					out:   []byte{123, 235},
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
				target:    []byte{123, 234},
			},
			mockHashCalls: []mockHashCall{
				{
					out:   []byte{5, 132, 99, 80, 1},
					times: 1,
				},
				{
					out:   []byte{35, 232, 59, 89, 1},
					times: 1,
				},
				{
					out:   []byte{123, 235},
					times: 1,
				},
				{
					out:   []byte{123, 236},
					times: 1,
				},
				{
					out:   []byte{123, 237},
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
				target:    []byte{123, 234},
			},
			mockHashCalls: []mockHashCall{
				{
					out:   []byte{123, 234},
					times: 1,
				},
				{
					out:   []byte{123, 233},
					times: 1,
				},
				{
					out:   []byte{0},
					times: 1,
				},
				{
					out:   []byte{5},
					times: 1,
				},
				{
					out:   []byte{25},
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
				func(lastCall bool, out []byte) {
					mockHasher.EXPECT().Hash(gomock.Any()).
						DoAndReturn(func(block *chain.Block) []byte {
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
		expected   struct {
			targetStr string
			hasErr    bool
		}
	}{
		{
			name:       "Target is maximum (easiest) if difficulty is minimum",
			difficulty: 1,
			expected: struct {
				targetStr string
				hasErr    bool
			}{
				targetStr: "115792089237316195423570985008687907853269984665640564039457584007913129639935",
				hasErr:    false,
			},
		},
		{
			name:       "Target is 1/1000000th of total range for a difficulty of 1000000",
			difficulty: 1000000,
			expected: struct {
				targetStr string
				hasErr    bool
			}{
				targetStr: "115792089237316195423570985008687907853269984665640564039457584007913129",
				hasErr:    false,
			},
		},
		{
			name:       "Uneven difficulty with significant fraction is taken into consideration",
			difficulty: 612325.9,
			expected: struct {
				targetStr string
				hasErr    bool
			}{
				targetStr: "189102060254704547073209780071556934681761644936635063270049858369936343",
				hasErr:    false,
			},
		},
		{
			name:       "Uneven difficulty with highly significant fraction is taken into consideration",
			difficulty: 1.9,
			expected: struct {
				targetStr string
				hasErr    bool
			}{
				targetStr: "60943204861745368861271012402782623032378921027243951008185249832298398445201",
				hasErr:    false,
			},
		},
		{
			name:       "Uneven and huge difficulty demonstrates math correctly",
			difficulty: 1743957183243897527420721548712409572190353046234907457134129375192.3417295172395673428,
			expected: struct {
				targetStr string
				hasErr    bool
			}{
				targetStr: "66396176666",
				hasErr:    false,
			},
		},
		{
			name:       "Trying to set target based on a difficulty <1 is an error",
			difficulty: 0.1,
			expected: struct {
				targetStr string
				hasErr    bool
			}{
				targetStr: "0",
				hasErr:    true,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := miner{}

			if err := m.SetTarget(c.difficulty); (err != nil) != c.expected.hasErr {
				t.Errorf("expected error: %v, got %v", c.expected.hasErr, err)
			}

			expectedTarget, ok := new(big.Int).SetString(c.expected.targetStr, 10)
			if !ok {
				t.Fatalf("Invalid argument to big int creation: %s", c.expected.targetStr)
			}

			if !bytes.Equal(m.target, expectedTarget.Bytes()) {
				t.Errorf("expected %v, got %v", expectedTarget.Bytes(), m.target)
			}
		})
	}
}
