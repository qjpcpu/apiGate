package conf

import (
	"github.com/gomodule/redigo/redis"
)

const (
	SESSION_ID   = "sessionid"
	COMM_USER_ID = "user-id"
)

var (
	g_conf     *Configure = &Configure{RedisConfig: &CacheConfig{}}
	g_cache    *redis.Pool
	g_dev_mode bool = false
)

func Get() *Configure {
	return g_conf
}

func Cache() *redis.Pool {
	return g_cache
}

func SetMode(isdev bool) {
	g_dev_mode = isdev
}

func IsDevMode() bool {
	return g_dev_mode
}
