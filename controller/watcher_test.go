package controller

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWatchRetry_retryConsul(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		maxRetry        int
		expectedRetries int
		breakLimit      int
	}{
		{
			name:            "happy path: retry 10x",
			maxRetry:        10,
			expectedRetries: 10,
			breakLimit:      20,
		},
		{
			name:            "happy path: retry 10x",
			maxRetry:        -1,
			breakLimit:      8,
			expectedRetries: 8,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wr := watcherRetry{
				maxRetries: tc.maxRetry,
				waitFunc: func(attempt uint, random *rand.Rand) int {
					return 1
				},
			}

			count := 0
			for true {
				isSuccessRetry, _ := wr.retryConsul(count)
				if !isSuccessRetry || count > tc.breakLimit {
					break
				}
				count++
			}

			assert.Equal(t, tc.expectedRetries, count-1)
		})
	}
}
