package myredis

import (
	"time"

	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
)

const (
	REDIS_T_HASH    = 1
	REDIS_T_SET     = 2
	REDIS_T_KEYS    = 3
	REDIS_T_STRING  = 4
	REDIS_T_LIST    = 5
	REDIS_T_SORTSET = 6
)

const (
	USER_INFO       = 1101 //用户信息
	USER_ID_LIST    = 1102 //用户列表
	USER_PHONE_LIST = 1103 //用户手机列表
	USER_EMAIL_LIST = 1103 //用户手机列表
	TOKEN_LIST      = 1104 //有效token
)

const (
	MGO_IDLE_COUNT   = 1   //连接池空闲个数
	MGO_ACTIVE_COUNT = 10  //连接池活动个数
	MGO_IDLE_TIMEOUT = 180 //空闲超时时间
)

var RedisClients map[int]*redis.Pool
var RedisQueue map[int]*redis.Pool
var RedisQueueKeys []string

func init() {
	str := beego.AppConfig.Strings("redis::addr")
	RedisClients = make(map[int]*redis.Pool)
	for index, value := range str {
		RedisClients[index] = createPool(MGO_IDLE_COUNT, MGO_ACTIVE_COUNT, MGO_IDLE_TIMEOUT, value)
	}

	str = beego.AppConfig.Strings("queue::addr")
	RedisQueue = make(map[int]*redis.Pool)
	for index, value := range str {
		RedisQueue[index] = createPool(MGO_IDLE_COUNT, MGO_ACTIVE_COUNT, MGO_IDLE_TIMEOUT, value)
	}

	str = beego.AppConfig.Strings("queue::keys")
	for _, value := range str {
		RedisQueueKeys = append(RedisQueueKeys, value)
	}
}

func createPool(maxIdle, maxActive, idleTimeout int, address string) (obj *redis.Pool) {
	obj = new(redis.Pool)
	obj.MaxIdle = maxIdle
	obj.MaxActive = maxActive
	obj.IdleTimeout = (time.Duration)(idleTimeout) * time.Second
	obj.Dial = func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", address)
		if err != nil {
			return nil, err
		}
		return c, err
	}
	return
}

func GetConn() (conn redis.Conn) {
	if len(RedisClients) <= 0 {
		return nil
	}

	return RedisClients[0].Get()
}

func GetQueueConn() (conn redis.Conn) {
	if len(RedisClients) <= 0 {
		return nil
	}

	return RedisClients[0].Get()
}
