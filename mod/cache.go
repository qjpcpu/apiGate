package mod

import (
	"github.com/garyburd/redigo/redis"
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

func (conf *CacheConfig) NewRedisPool() *redis.Pool {
	conf.setDefault()
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

func (cc *CacheConfig) setDefault() {
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
