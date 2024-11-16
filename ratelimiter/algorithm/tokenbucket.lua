-- Keys and args
local key = KEYS[1]
local timestamp = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local emissionRate = tonumber(ARGV[3])
local emissionAmount = tonumber(ARGV[4])

-- Get the bucket values of token and last refresh
local tokens = redis.call("HGET", key, "tokens")
local lastRefresh = redis.call("HGET", key, "lastRefresh")

-- If either is missing, set the default values
if tokens == false or lastRefresh == false then
	tokens = capacity
	lastRefresh = timestamp
end

-- Calculate the number of tokens in the bucket
local delta = timestamp - lastRefresh
local ticks = math.floor(delta / emissionRate)
local newTokens = math.floor(ticks * emissionAmount)
local currentTokens = math.min(capacity, newTokens + tokens)

-- Return early if we don't have any tokens
if currentTokens <= 0 then
	return 0
end

-- Remove a token from the bucket as request is allowed
currentTokens = currentTokens - 1

-- Set the refreshed timestamp to the number of ticks
local refreshed = lastRefresh + math.floor(ticks * emissionRate)

-- Set the new bucket params
redis.call("HSET", key, "tokens", currentTokens, "lastRefresh", refreshed)

-- Calculate the total time to fill the bucket
local period = capacity / emissionAmount * emissionRate

-- Refresh the expiration on the users bucket to 10x the period
redis.call("EXPIRE", key, period * 10)

return 1
