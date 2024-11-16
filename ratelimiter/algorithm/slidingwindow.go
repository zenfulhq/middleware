package algorithm

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed slidingwindow.lua
var script string

type SlidingWindow struct {
	// Algorithm Configuration
	period time.Duration // Time period for the rate
	limit  int64         // Request limit for the period

	// Dependencies
	client *redis.Client // KV Store client
	script *redis.Script // Lua script for the algorithm
}

func NewSlidingWindow(
	limit int64,
	period time.Duration,
	client *redis.Client,
) (*SlidingWindow, error) {
	return &SlidingWindow{
		period: period,
		limit:  limit,
		client: client,
		script: redis.NewScript(string(script)),
	}, nil
}

func (a *SlidingWindow) IsAllowed(ctx context.Context, clientID string) (bool, error) {
	// Derive key from clientID
	key := fmt.Sprintf("window:%s", clientID)

	// Generate the current timestamp in microseconds
	now := time.Now().UnixMicro()

	// Calculate the cutoff
	cutoff := now - a.period.Microseconds()

	// Run the script with the key and args
	res := a.script.Run(
		ctx, a.client, []string{key}, now, cutoff, a.period.Seconds(), a.limit,
	)

	return res.Bool()
}
