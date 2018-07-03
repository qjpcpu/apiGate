package conf

import (
	"github.com/garyburd/redigo/redis"
	"github.com/qjpcpu/apiGate/mod"
	"github.com/qjpcpu/apiGate/unique-id"
	"github.com/qjpcpu/log"
	"os"
)

func InitCache() {
	g_cache = Get().RedisConfig.NewRedisPool()
}

func InitIDGenerator(redisPool *redis.Pool) {
	conn := redisPool.Get()
	defer conn.Close()
	script := redis.NewScript(1, `
local offset=tonumber(ARGV[1]) + 1
local serv_id=tonumber(redis.call("INCRBY",KEYS[1],offset))
if serv_id>=1024 then
 serv_id = offset
 redis.call("SET",KEYS[1],offset)
end
return serv_id
`)
	idWorkerNum := 5
	seed, err := redis.Int(script.Do(conn, "apigate_id_seed", idWorkerNum))
	if err != nil {
		log.Errorf("conn to redis failed:%v", err)
		os.Exit(1)
	}
	uid.InitGenerator(seed-idWorkerNum, idWorkerNum)
}

func GetFreqCtrl(threshold, window int64) *mod.FreqController {
	fc, _ := mod.NewFreqController(g_cache, threshold, window)
	return fc
}
