package algorithm

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBucket struct {
	// Algorithm Configuration
	capacity       int           // The max capacity of the bucket
	emissionRate   time.Duration // Period in which a token is added to bucket
	emissionAmount int           // Amount of tokens added to the bucket per tick

	// Dependencies
	client *redis.Client // KV Store client
	script *redis.Script // Lua script for the algorithm
}

func NewTokenBucket(
	capacity int,
	rate time.Duration,
	amount int,
	client *redis.Client,
) (*TokenBucket, error) {
	loc, exists := os.LookupEnv("SCRIPT_DIR")
	if !exists {
		loc = "../algorithm"
	}

	script, err := os.ReadFile(fmt.Sprintf("%s/tokenbucket.lua", loc))
	if err != nil {
		return nil, fmt.Errorf("failed to read script: %w", err)
	}

	return &TokenBucket{
		capacity:       capacity,
		emissionRate:   rate,
		emissionAmount: amount,
		client:         client,
		script:         redis.NewScript(string(script)),
	}, nil
}

func (a *TokenBucket) IsAllowed(ctx context.Context, clientID string) (bool, error) {
	// Derive key from clientID
	key := fmt.Sprintf("bucket:%s", clientID)

	// Generate the current timestamp in microseconds
	now := time.Now().UnixNano()

	// Run the script with the key and args
	res := a.script.Run(
		ctx, a.client, []string{key}, now, a.capacity, a.emissionRate, a.emissionAmount,
	)

	return res.Bool()
}
