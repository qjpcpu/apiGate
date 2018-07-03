package mod

import (
	"github.com/garyburd/redigo/redis"
	"strings"
	"time"
)

type CacheConfig struct {
	RedisNetwork        string `json:"network" yaml:"network"`
	RedisAddress        string `json:"addr" yaml:"addr"`
	RedisPassword       string `json:"password" yaml:"password"`
	RedisConnectTimeout int    `json:"conn_timeout" yaml:"conn_timeout"`
	RedisReadTimeout    int    `json:"read_timeout" yaml:"read_timeout"`
	RedisWriteTimeout   int    `json:"write_timeout" yaml:"write_timeout"`
	RedisMaxActive      int    `json:"max_active" yaml:"max_active"`
	RedisMaxIdle        int    `json:"max_idle" yaml:"max_idle"`
	RedisIdleTimeout    int    `json:"idle_timeout" yaml:"idle_timeout"`
	RedisWait           bool   `json:"wait" yaml:"wait"`
	RedisDb             string `json:"db_num" yaml:"db_num"`
}

type Cache struct {
	Conf      CacheConfig
	redisPool *redis.Pool
}

const (
	Success int = iota + 1
	KeyNotFound
	RedisError
)

func CheckRedisReturnValue(err error) int {
	if err != nil && strings.Contains(err.Error(), "nil returned") {
		return KeyNotFound
	} else if err == nil {
		return Success
	} else {
		return RedisError
	}
}

func (cache *Cache) RedisPool() *redis.Pool {
	if cache.redisPool == nil {
		cache.redisPool = cache.newRedisPool(&cache.Conf)
	}
	return cache.redisPool
}

func (cache *Cache) newRedisPool(conf *CacheConfig) *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {

			var connect_timeout time.Duration = time.Duration(conf.RedisConnectTimeout) * time.Second
			var read_timeout = time.Duration(conf.RedisReadTimeout) * time.Second
			var write_timeout = time.Duration(conf.RedisWriteTimeout) * time.Second

			c, err := redis.DialTimeout(conf.RedisNetwork, conf.RedisAddress, connect_timeout, read_timeout, write_timeout)

			if err != nil {
				return nil, err
			}
			if len(conf.RedisPassword) > 0 {
				if _, err := c.Do("AUTH", conf.RedisPassword); err != nil {
					c.Close()
					return nil, err
				}
			}

			if len(conf.RedisDb) > 0 {
				if _, err = c.Do("SELECT", conf.RedisDb); err != nil {
					c.Close()
					return nil, err
				}
			}

			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			// redis 集群不使用ping命令
			//_, err := c.Do("PING")
			return nil
		},
		MaxIdle:     conf.RedisMaxIdle,
		MaxActive:   conf.RedisMaxActive,
		IdleTimeout: time.Duration(conf.RedisIdleTimeout) * time.Second,
		Wait:        conf.RedisWait,
	}
}

func (cache *Cache) GetEXBytes(key string, timeout int) ([]byte, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	bytes, err := redis.Bytes(conn.Do("GET", key))
	if err == nil && bytes != nil {
		conn.Do("EXPIRE", key, timeout)
	}
	return bytes, err
}

func (cache *Cache) Incr(key string) (int, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.Int(conn.Do("INCR", key))

	return res, err
}

func (cache *Cache) MGet(key []interface{}) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := conn.Do("MGET", key...)
	return res, err
}

func (cache *Cache) MGetValue(keys []interface{}) ([]interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.Values(conn.Do("MGET", keys...))
	return res, err
}

func (cache *Cache) Get(key string) ([]byte, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()

	res, err := redis.Bytes(conn.Do("GET", key))
	return res, err
}

func (cache *Cache) GetString(key string) (string, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	return redis.String(conn.Do("GET", key))
}

func (cache *Cache) MSET(keys ...interface{}) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("MSET", keys...)
	return err
}

func (cache *Cache) SetEXBytes(key string, bytes []byte, timeout int) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, bytes, "EX", timeout)
	return err
}

func (cache *Cache) Set(key string, bytes []byte) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, bytes)
	return err
}

func (cache *Cache) ZrevrangeByScore(key string, max_num, min_num int, withscores bool, offset, count int) ([]int, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	var res []int
	var err error
	if !withscores {
		res, err = redis.Ints(conn.Do("ZREVRANGEBYSCORE", key, max_num, min_num, "limit", offset, count))
	} else {
		res, err = redis.Ints(conn.Do("ZREVRANGEBYSCORE", key, max_num, min_num, "withscores", "limit", offset, count))
	}
	return res, err
}

func (cache *Cache) ZrangeByScore(key string, min_num, max_num int64, withscores bool, offset, count int) ([]int, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	var res []int
	var err error
	if withscores {
		res, err = redis.Ints(conn.Do("ZREVRANGEBYSCORE", key, max_num, min_num, "limit", offset, count))
	} else {
		res, err = redis.Ints(conn.Do("ZREVRANGEBYSCORE", key, max_num, min_num, "withscores", "limit", offset, count))
	}
	return res, err
}

func (cache *Cache) Zrange(key string, start, end int, withscores bool) ([]string, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	var res []string
	var err error
	if withscores {
		res, err = redis.Strings(conn.Do("ZRANGE", key, start, end, "withscores"))
	} else {
		res, err = redis.Strings(conn.Do("ZRANGE", key, start, end))
	}
	return res, err
}

func (cache *Cache) ZrangeInts(key string, start, end int, withscores bool) ([]int, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	var res []int
	var err error
	if withscores {
		res, err = redis.Ints(conn.Do("ZRANGE", key, start, end, "withscores"))
	} else {
		res, err = redis.Ints(conn.Do("ZRANGE", key, start, end))
	}
	return res, err
}

func (cache *Cache) SETEX(key string, val interface{}, timeout int) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("SETEX", key, timeout, val)
	return err
}

func (cache *Cache) INCRBY(key string, count int64, timeout int) (int64, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	n, err := redis.Int64(conn.Do("INCRBY", key, count))
	if err == nil && timeout >= 0 {
		conn.Do("EXPIRE", key, timeout)
	}
	return n, err
}

func (cache *Cache) Zscore(key, member string) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()

	res, err := conn.Do("ZSCORE", key, member)
	return res, err
}

func (cache *Cache) Zadd(key string, value, member interface{}) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()

	res, err := conn.Do("ZADD", key, value, member)
	return res, err
}

func (cache *Cache) SADD(key string, items []string, timeout int) error {
	var err error
	conn := cache.RedisPool().Get()
	defer conn.Close()
	if timeout <= 0 {
		_, err = conn.Do("SADD", key, items)
	} else {
		_, err = conn.Do("SADD", key, items, timeout)
	}

	return err

}

func (cache *Cache) TYPE(key string) string {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	type_str, err := conn.Do("TYPE", key)
	if err != nil {
		return ""
	}

	return type_str.(string)
}

func (cache *Cache) SISMEMBER(key string, bytes []byte) bool {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	exists, err := redis.Bool(conn.Do("SISMEMBER", key, bytes))
	if err != nil {
		return false
	}

	if exists {
		return true
	} else {
		return false
	}

}

func (cache *Cache) EXISTS(key string) bool {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false
	}

	if exists {
		return true
	} else {
		return false
	}
}

func (cache *Cache) HINCRBY(key, field string, count int64) (int64, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	return redis.Int64(conn.Do("HINCRBY", key, field, count))
}

func (cache *Cache) Hmget(key string, fields []string) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()

	var args []interface{}
	args = append(args, key)
	for _, field := range fields {
		args = append(args, field)
	}

	res, err := conn.Do("HMGET", args...)

	return res, err
}

func (cache *Cache) GetStringMap(key string) (map[string]string, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.StringMap(conn.Do("HGETALL", key))
	return res, err
}

func (cache *Cache) HGetAll(key string) ([]byte, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.Bytes(conn.Do("HGETALL", key))

	return res, err
}

func (cache *Cache) HSET(key, field string, data interface{}) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("HSET", key, field, data)
	return err
}

func (cache *Cache) HMset(value []interface{}) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := conn.Do("HMSET", value...)

	return res, err
}

func (cache *Cache) HGET(key, field string) (string, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	return redis.String(conn.Do("HGET", key, field))
}

func (cache *Cache) HDEL(key, field string) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("HDEL", key, field)
	return err
}

func (cache *Cache) DEL(key string) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}

func (cache *Cache) Expire(key string, timeout int) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("EXPIRE", key, timeout)

	return err
}

func (cache *Cache) HIncrby(key, field string, value interface{}) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := conn.Do("HINCRBY", key, field, value)

	return res, err
}

func (cache *Cache) Rpush(key string, value interface{}) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("RPUSH", key, value)
	return err
}

func (cache *Cache) RpushBatch(keys []interface{}) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("RPUSH", keys...)
	return err
}

func (cache *Cache) Lrange(key string, start, end int) ([]interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	result, err := redis.Values(conn.Do("LRANGE", key, start, end))
	return result, err
}

func (cache *Cache) Lrem(key string, value interface{}) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("LREM", key, 0, value)
	return err
}

func CreateCache(conf *CacheConfig) (*Cache, error) {
	cache := Cache{}
	cache.Conf = *conf
	pool := cache.RedisPool()
	err := pool.TestOnBorrow(pool.Get(), time.Now())
	if err != nil {
		return nil, err
	}

	return &cache, nil
}

func NewCacheConfig() *CacheConfig {
	cc := &CacheConfig{}
	cc.SetDefault()
	return cc
}

func (cc *CacheConfig) SetDefault() {
	if cc.RedisNetwork == "" {
		cc.RedisNetwork = "tcp"
	}
	if cc.RedisAddress == "" {
		cc.RedisAddress = "127.0.0.1:6379"
	}
	if cc.RedisConnectTimeout == 0 {
		cc.RedisConnectTimeout = 2
	}
	if cc.RedisReadTimeout == 0 {
		cc.RedisReadTimeout = 2
	}
	if cc.RedisWriteTimeout == 0 {
		cc.RedisWriteTimeout = 2
	}
	if cc.RedisMaxIdle == 0 {
		cc.RedisMaxIdle = 50
	}
	if cc.RedisIdleTimeout == 0 {
		cc.RedisIdleTimeout = 600
	}
	cc.RedisWait = true
}
