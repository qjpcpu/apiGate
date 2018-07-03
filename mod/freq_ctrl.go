package mod

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

const (
	FreqCtrlKey = "freqctrlhash"
)

// 用户频控漏洞算法redis实现:
// 存储基本数据对象为hash,hash_name是user_id,hash的field是url
// hash的field_value是一个json结构:
// {
//  "of": 12,            // OverFlow,溢出令牌数量,记录超发的令牌数量
//  "last": 3245345234,  // 上次访问该规则的时间戳,该时间戳为unix nano
//  "t": 30,             // 时间窗口内可用令牌数量,初始值即为频控最大值
// }
// 频控时间片默认10ms
// 1.每次访问该数据结构,将可用令牌数"回血": (当期时间戳 - 上次访问该规则的时间戳) / 时间片大小 * 令牌回血率, 令牌回血率 = 频控阈值 / (时间窗大小 / 时间片大小)
// 2. 如果是tick(is_tick=1)操作,将令牌数t-1,如果t==0,则将溢出令牌数记到of值,实际上目前的算法是没有超发的,记录溢出仅仅是为了反馈调用的溢出水位百分比
// 3. 返回当前频率计数值threshold - t - of,由于of可能<0,故返回计数>=threshold,反映出溢出水位
// redis-cli --eval tick.lua hash_name , field_name freq_threshold freq_window  timestamp is_tick
var tickScript = redis.NewScript(1, `
local key = KEYS[1]
local field = ARGV[1]
local threshold = tonumber(ARGV[2])
local time_window = tonumber(ARGV[3]) * 1000
local timestamp = tonumber(ARGV[4])
local interval = 10
local is_tick = tonumber(ARGV[5])
if redis.call("HEXISTS", key,field) == 0 then
  if is_tick ~= 1 then return 0 end
  redis.call("HSET",key,field,string.format('{"of":0,"last":%d,"t":%d}',timestamp,threshold-1))
  redis.call("EXPIRE",key,time_window/1000)
  return 1
end
local data = cjson.decode(redis.call("HGET",key,field))
data['t'] = data['t'] + (timestamp - data['last'])*1000/interval*(threshold*interval/time_window)
if data['t'] > threshold then data['t'] = threshold end
if is_tick == 1 then data['t'] = data['t'] - 1 end
if data['t'] < 0 then
  data['of'] = data['t']
  data['t'] = 0
elseif data['t'] >= 1 then
  data['of'] = 0
end
data['last'] = timestamp
redis.call("HSET",key,field,cjson.encode(data))
redis.call("EXPIRE",key,time_window/1000)
return threshold - data['t'] - data['of']
`)

type FreqController struct {
	pool      *redis.Pool
	threshold int64
	window    int64 // freq control time window(seconds)
}

// NewFreqController创建频控对象,设置时间窗口win(秒)内最大次数thr
func NewFreqController(p *redis.Pool, thr, win int64) (*FreqController, error) {
	if thr < 1 || win < 1 {
		return nil, errors.New("error parameters")
	}
	return &FreqController{pool: p, threshold: thr, window: win}, nil
}

// Tick频控次数+1并返回频控当前水位[0,1.0], >1表示超过阈值
func (fc *FreqController) Tick(user, rule string) float64 {
	if user == "" || rule == "" {
		return 0
	}
	conn := fc.pool.Get()
	defer conn.Close()
	fcnt, err := redis.Int64(tickScript.Do(conn, fc.key(user), rule, fc.threshold, fc.window, time.Now().Unix(), 1))
	if err != nil {
		return 0
	}
	return float64(fcnt) / float64(fc.threshold)
}

// Check频控次数,返回频控当前水位[0,1.0], >1表示超过阈值
func (fc *FreqController) Check(user, rule string) float64 {
	if user == "" || rule == "" {
		return 0
	}
	conn := fc.pool.Get()
	defer conn.Close()
	fcnt, err := redis.Int64(tickScript.Do(conn, fc.key(user), rule, fc.threshold, fc.window, time.Now().Unix(), 0))
	if err != nil {
		return 0
	}
	return float64(fcnt) / float64(fc.threshold)
}

func (fc *FreqController) key(k string) string {
	return fmt.Sprintf("%s:%v|%v:%s", FreqCtrlKey, fc.threshold, fc.window, k)
}
