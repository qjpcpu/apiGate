package middlewares

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"github.com/qjpcpu/apiGate/conf"
)

func CacheUser(sid string, uid string) error {
	conn := conf.Cache().Get()
	defer conn.Close()
	_, err := conn.Do("SET", sid, []byte(uid), "EX", conf.Get().SessionExpireSeconds)
	return err
}

func RemoveUser(sid string) error {
	conn := conf.Cache().Get()
	defer conn.Close()
	_, err := conn.Do("DEL", sid)
	return err
}

func FetchUser(key string) (string, error) {
	conn := conf.Cache().Get()
	defer conn.Close()
	user_id, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return "", err
	}
	if user_id == "" {
		return "", errors.New("invalid user id from cache")
	}
	return user_id, nil
}
