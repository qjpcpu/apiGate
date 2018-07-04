package conf

import (
	"github.com/garyburd/redigo/redis"
)

const (
	SESSION_ID   = "sessionid"
	COMM_USER_ID = "user-id"
)

var (
	g_conf  *Configure = &Configure{}
	g_cache *redis.Pool
)

func Get() *Configure {
	return g_conf
}

func Cache() *redis.Pool {
	return g_cache
}
