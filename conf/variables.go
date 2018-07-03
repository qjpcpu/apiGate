package conf

import (
	"github.com/garyburd/redigo/redis"
	"github.com/qjpcpu/apiGate/mod"
)

const (
	SESSION_ID   = "sessionid"
	COMM_USER_ID = "user_id"
)

var (
	g_conf  *mod.Configure = &mod.Configure{}
	g_cache *redis.Pool
)

func Get() *mod.Configure {
	return g_conf
}

func Cache() *redis.Pool {
	return g_cache
}
