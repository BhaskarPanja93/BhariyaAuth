package stores

import (
	"context"
	"fmt"
	"time"

	Secrets "BhariyaAuth/constants/secrets"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	Ctx         context.Context
)

func ConnectRedis() {
	if RedisClient != nil {
		return
	}
	Ctx = context.Background()

	useSocket := true
	for {
		var (
			client *redis.Client
			err    error
		)

		if useSocket {
			fmt.Println("Trying Redis via UNIX socket...")
			client = redis.NewClient(&redis.Options{
				Network:      "unix",
				Addr:         Secrets.RedisSocket,
				Password:     Secrets.RedisPassword,
				DB:           0,
				MinIdleConns: 10,
				PoolSize:     2000,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
			})
		} else {
			fmt.Println("Trying Redis via TCP/IP...")
			client = redis.NewClient(&redis.Options{
				Network:      "tcp",
				Addr:         Secrets.RedisHost + ":" + Secrets.RedisPort,
				Password:     Secrets.RedisPassword,
				DB:           0,
				MinIdleConns: 10,
				PoolSize:     2000,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
			})
		}

		_, err = client.Ping(Ctx).Result()
		if err != nil {
			fmt.Println("Redis connection failed", err.Error())
			_ = client.Close()
			useSocket = !useSocket
			time.Sleep(2 * time.Second)
			continue
		}

		RedisClient = client
		break
	}

	fmt.Println("Redis connection established successfully")
}
