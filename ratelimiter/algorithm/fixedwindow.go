package algorithm

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type FixedWindow struct {
	// Algorithm Configuration
	period time.Duration // Time period for the rate
	limit  int64         // Request limit for the period

	// Dependencies
	client *redis.Client // KV store client
}

func NewFixedWindow(limit int64, period time.Duration, client *redis.Client) (*FixedWindow, error) {
	// Ensure the time period is at least 1 second.
	if period < time.Second {
		period = time.Second
	}

	// Constrain the time period to seconds granularity.
	period = period / time.Second * time.Second

	return &FixedWindow{
		period: period,
		limit:  limit,
		client: client,
	}, nil
}

func (a *FixedWindow) IsAllowed(ctx context.Context, clientID string) (bool, error) {
	// Calculate the current time window truncating to the bucket floor.
	currentWindow := time.Now().Truncate(a.period).Unix()

	// Convert current time window into a string for the bucket key.
	win := strconv.FormatInt(currentWindow, 10)
	key := fmt.Sprintf("%s:%s", win, clientID)

	// Create a new pipline to multi operations
	pipeline := a.client.Pipeline()

	// Incr the client ID in the current window by 1
	incr := pipeline.Incr(ctx, key)

	// Expire the bucket if it does not have an expiry set.
	pipeline.ExpireNX(ctx, key, a.period*2)

	// Execute the pipeline
	pipeline.Exec(ctx)

	// Obtain the value from the incr command and check if request is allowed.
	isAllowed := a.limit >= incr.Val()

	return isAllowed, nil
}
