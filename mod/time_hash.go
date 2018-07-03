package mod

import (
	"encoding/json"
	"errors"
	"github.com/garyburd/redigo/redis"
	"time"
)

/*
 数据存储的样例: key: {"data":YOUR_DATA,"comm":{"exp":1487311773,"ver": 2}}
 data存放用户数据,comm存放内置数据
 comm.ver: 每SET一次数据,ver+1
 comm.exp: 过期unix时间戳
*/
const (
	CacheEngineKey = "wm:cache-engine"
)

var ErrNoCache = errors.New("No such data")

// redis-cli --eval h.lua h , c current_timestamp  isCut
var getScript = redis.NewScript(1, `
if redis.call("HEXISTS", KEYS[1],ARGV[1]) == 0 then
  return nil
end
local payload = redis.call("HGET",KEYS[1],ARGV[1])
local data = cjson.decode(payload)
if tonumber(data["comm"]["exp"]) < tonumber(ARGV[2]) then
  redis.call("HDEL",KEYS[1],ARGV[1])
  return nil
else
  if tonumber(ARGV[3]) == 1 then
    redis.call("HDEL",KEYS[1],ARGV[1])
  end
  return payload
end
`)

// redis-cli --eval h.lua h , c  'data' current_timestamp
var setScript = redis.NewScript(1, `
if redis.call("HEXISTS", KEYS[1],ARGV[1]) == 1 then
  local data = cjson.decode(ARGV[2])
  local old_data = cjson.decode(redis.call("HGET",KEYS[1],ARGV[1]))
  if old_data["comm"]["ver"] == nil then old_data["comm"]["ver"] = 0 end
  if old_data["comm"]["exp"] == nil or tonumber(old_data["comm"]["exp"]) < tonumber(ARGV[3]) then old_data["comm"]["ver"] = -1 end
  data["comm"]["ver"] = old_data["comm"]["ver"] + 1
  ARGV[2] = cjson.encode(data)
end
redis.call("HSET",KEYS[1],ARGV[1],ARGV[2])
local decoded = cjson.decode(ARGV[2])
local exp = tonumber(decoded["comm"]["exp"])
local exp_key = "__expire_at_max__"
if redis.call("HEXISTS", KEYS[1],exp_key) == 1 then
  local oexp = tonumber(redis.call("HGET",KEYS[1],exp_key))
  if oexp < exp then
    redis.call("HSET",KEYS[1],exp_key,exp)
    redis.call("EXPIRE",KEYS[1],exp - ARGV[3])
  end
else
  redis.call("HSET",KEYS[1],exp_key,exp)
  redis.call("EXPIRE",KEYS[1],exp - ARGV[3])
end
return tonumber(decoded["comm"]["ver"])
`)

// reids-cli --eval h.lua KEY_NAME , current_timestamp  scan_count
var scanScript = redis.NewScript(1, `
	local cur_time = tonumber(ARGV[1])
  local expire_info=redis.call("ZRANGE", KEYS[1],0,tonumber(ARGV[2]),"WITHSCORES")
	local index = 1
	local max = table.getn(expire_info)
  local vals = {}
	while (index + 1) <= max
  do
		local member = expire_info[index]
		local expire_time = expire_info[index+1] + 0
		index = index + 2
		if expire_time > cur_time then
			break
		end
    table.insert(vals, member)
		redis.call("ZREM", KEYS[1], member)
  end
  return vals
`)

// redis-cli --eval h.lua h ,  current_timestamp f1 f2 ...
var cleanScript = redis.NewScript(1, `
local dels = 0
local tm = tonumber(ARGV[1])
for i,v in ipairs(ARGV) do
  if i > 1 and redis.call("HEXISTS", KEYS[1],v) == 1 then
    local payload = redis.call("HGET",KEYS[1],v)
    local data = cjson.decode(payload)
    if tonumber(data["comm"]["exp"]) < tm then
      redis.call("HDEL",KEYS[1],v)
      dels = dels + 1
    end
  end
end
return dels
`)

type CacheEngine struct {
	pool *redis.Pool
	Name string
}

type CommonPayload struct {
	ExpireAt int64 `json:"exp"`
	Version  int64 `json:"ver"`
}

type cachePayloadIn struct {
	CommonPayload `json:"comm"`
	Data          interface{} `json:"data"`
}

type cachePayloadOut struct {
	CommonPayload `json:"comm"`
	Data          *json.RawMessage `json:"data"`
}

func NewCacheEngine(p *redis.Pool, names ...string) *CacheEngine {
	ce := &CacheEngine{pool: p}
	if len(names) > 0 {
		ce.Name = names[0]
	}
	return ce
}

func (ce *CacheEngine) getCEKey() string {
	if ce.Name != "" {
		return CacheEngineKey + ":" + ce.Name
	} else {
		return CacheEngineKey
	}
}

func (ce *CacheEngine) getExpKey() string {
	return ce.getCEKey() + ":exp"
}

func (ce *CacheEngine) Del(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	conn := ce.pool.Get()
	defer conn.Close()
	params := []interface{}{ce.getCEKey()}
	for _, key := range keys {
		params = append(params, key)
	}
	_, err := conn.Do("HDEL", params...)
	return err
}

func (ce *CacheEngine) Clean(counts ...int) (int, error) {
	count := 100
	if len(counts) == 1 && counts[0] > 0 && counts[0] <= 1000 {
		count = counts[0]
	}
	var dels, d int
	var err error
	for {
		d, err = ce.CleanOnce(count)
		dels += d
		if err != nil {
			break
		}
		if d <= 0 {
			break
		}
	}
	return dels, err
}

func (ce *CacheEngine) CleanOnce(count int) (int, error) {
	conn := ce.pool.Get()
	defer conn.Close()
	now := time.Now().Unix()
	list, err := redis.Strings(scanScript.Do(conn, ce.getExpKey(), now, count))
	if err != nil {
		return 0, err
	}
	if len(list) == 0 {
		return 0, nil
	}
	params := []interface{}{ce.getCEKey(), now}
	for _, item := range list {
		params = append(params, item)
	}
	dels, _ := redis.Int(cleanScript.Do(conn, params...))
	return dels, nil
}

// Put 放入数据并返回数据版本号
func (ce *CacheEngine) Put(key string, data interface{}, expirtAts ...time.Time) (int64, error) {
	now := time.Now()
	var expirtAt time.Time
	if len(expirtAts) > 0 {
		if expirtAts[0].Before(now) {
			return 0, nil
		}
		expirtAt = expirtAts[0]
	} else {
		expirtAt = time.Now().AddDate(1, 0, 0)
	}
	payload := cachePayloadIn{
		Data: data,
		CommonPayload: CommonPayload{
			ExpireAt: expirtAt.Unix(),
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	conn := ce.pool.Get()
	defer conn.Close()
	ver, err := redis.Int64(setScript.Do(conn, ce.getCEKey(), key, b, time.Now().Unix()))
	if err == nil {
		conn.Do("ZADD", ce.getExpKey(), expirtAt.Unix(), key)
	}
	return ver, err
}

// data 必须为指针
func (ce *CacheEngine) Get(key string, data interface{}, isCut ...bool) (int64, error) {
	conn := ce.pool.Get()
	defer conn.Close()
	cut := 0
	if len(isCut) == 1 && isCut[0] {
		cut = 1
	}
	res, err := redis.Bytes(getScript.Do(conn, ce.getCEKey(), key, time.Now().Unix(), cut))
	if err != nil {
		if err == redis.ErrNil {
			return 0, ErrNoCache
		}
		return 0, err
	}
	p := cachePayloadOut{}
	if err = json.Unmarshal(res, &p); err != nil {
		return 0, err
	}
	return p.CommonPayload.Version, json.Unmarshal(*p.Data, data)
}
