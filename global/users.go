package global

import (
	"errors"
)

func CacheUser(sid string, uid string) error {
	err := G_cache.SetEXBytes(sid, []byte(uid), G.SessionExpireSeconds)
	if err != nil {
		return err
	}
	return nil
}

func RemoveUser(sid string) error {
	return G_cache.DEL(sid)
}

type _userInfo struct {
	UserId string
}

func FetchUser(key string) (string, error) {
	user_id, err := G_cache.GetString(key)
	if err != nil {
		return "", err
	}
	if user_id == "" {
		return "", errors.New("invalid user id from cache")
	}
	return user_id, nil
}
