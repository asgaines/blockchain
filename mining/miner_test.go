package mining

import (
	"testing"
	"time"
)

func TestGetDifficulty(t *testing.T) {
	cases := []struct {
		message        string
		miner          miner
		actualDur      time.Duration
		currDifficulty float64
		expected       float64
	}{
		{
			message: "An exact match between actual and desired duration returns the same difficulty",
			miner: miner{
				targetDurPerBlock: 10 * time.Minute,
			},
			actualDur:      10 * time.Minute,
			currDifficulty: 1024,
			expected:       1024,
		},
		{
			message: "An actual duration half of expected returns a difficulty twice of the current value",
			miner: miner{
				targetDurPerBlock: 10 * time.Minute,
			},
			actualDur:      5 * time.Minute,
			currDifficulty: 1024,
			expected:       2048,
		},
		{
			message: "An actual duration twice of expected returns a difficulty half of the current value",
			miner: miner{
				targetDurPerBlock: 10 * time.Minute,
			},
			actualDur:      20 * time.Minute,
			currDifficulty: 1024,
			expected:       512,
		},
		{
			message: "An actual duration 1.5 times of expected returns a difficulty quotient of 1.5 of the current value",
			miner: miner{
				targetDurPerBlock: 10 * time.Minute,
			},
			actualDur:      15 * time.Minute,
			currDifficulty: 1024,
			expected:       1024 / 1.5,
		},
	}

	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			got := c.miner.calcDifficulty(c.actualDur, c.currDifficulty)

			if got != c.expected {
				t.Errorf("expected %v, got %v", c.expected, got)
			}
		})
	}
}
