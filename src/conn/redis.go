package conn

import (
	"log"

	"github.com/go-redis/redis"
)

var client *redis.Client

func init() {
	// 创建一个新的Redis客户端
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// 检查连接
	pong, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
	log.Printf("connected to redis and ping get %v", pong)
}

func SetRedisKV(k string, v string) error {
	return client.Set(k, v, 0).Err()
}

func GetRedisVal(k string) (string, error) {
	return client.Get(k).Result()
}

func DelRedisKey(k string) error {
	return client.Del(k).Err()
}
