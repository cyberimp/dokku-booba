local set = KEYS[1]
local list = KEYS[2]
local newID = ARGV[1]
local trim = tonumber(ARGV[2])

redis.call('SADD', set, newID)
local length = tonumber(redis.call('LPUSH', list, newID))
if length > trim then
    local diff = length-trim
    local result = redis.call('LRANGE', list, -diff, -1)
    for i in result do
        redis.call('SREM', set, i)
    end
    redis.call('LTRIM', list, 0, trim-1)
end

return 1
