package stores

import (
	Config "BhariyaAuth/constants/config"
	"fmt"
	"time"

	Secrets "BhariyaAuth/constants/secrets"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis() {
	if RedisClient != nil {
		return
	}

	var useSocket = true

	for {
		if useSocket {
			fmt.Println("Trying Redis via UNIX socket...")
			RedisClient = redis.NewClient(&redis.Options{
				Network:      "unix",
				Addr:         Secrets.RedisSocket,
				Password:     Secrets.RedisPassword,
				DB:           0,
				MinIdleConns: 1,
				PoolSize:     50,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
			})
		} else {
			fmt.Println("Trying Redis via TCP/IP...")
			RedisClient = redis.NewClient(&redis.Options{
				Network:      "tcp",
				Addr:         Secrets.RedisHost + ":" + Secrets.RedisPort,
				Password:     Secrets.RedisPassword,
				DB:           0,
				MinIdleConns: 1,
				PoolSize:     20,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
			})
		}

		_, err := RedisClient.Ping(Config.CtxBG).Result()
		if err != nil {
			fmt.Println("Redis connection failed", err.Error())
			_ = RedisClient.Close()
			useSocket = !useSocket
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	fmt.Println("Redis connection established successfully")
}
