-- Keys and args
local key = KEYS[1]
local timestamp = tonumber(ARGV[1])
local cutoff = tonumber(ARGV[2])
local period = tonumber(ARGV[3])
local limit = tonumber(ARGV[4])

-- Remove any expired events from the sorted set
redis.call("ZREMRANGEBYSCORE", key, "-inf", cutoff)

-- Refresh the expiration on the set to 10x the period
redis.call("EXPIRE", key, period * 10)

-- Get the count of events in the current window
local count = redis.call("ZCOUNT", key, "-inf", "inf")

-- If the count is greater than the limit, return early
if count >= limit then
	return 0
end

-- Add in the current event to the sorted set
redis.call("ZADD", key, timestamp, timestamp)

return 1
