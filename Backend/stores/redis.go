package stores

import (
	Config "BhariyaAuth/constants/config"
	Logs "BhariyaAuth/processors/logs"
	"time"

	Secrets "BhariyaAuth/constants/secrets"

	"github.com/redis/go-redis/v9"
)

const redisFileName = "stores/redis"

var RedisClient *redis.Client

func ConnectRedis() {
	Logs.RootLogger.Add(Logs.Intent, redisFileName, "", "Attempting connect Redis")

	if RedisClient != nil {
		return
	}

	var useSocket = true

	for {

		if useSocket {
			Logs.RootLogger.Add(Logs.Intent, redisFileName, "", "Redis using unix socket")

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
			Logs.RootLogger.Add(Logs.Intent, redisFileName, "", "Redis using TCP")

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
			Logs.RootLogger.Add(Logs.Error, redisFileName, "", "Redis ping error: "+err.Error())

			_ = RedisClient.Close()

			useSocket = !useSocket

			time.Sleep(2 * time.Second)

			continue
		}
		Logs.RootLogger.Add(Logs.Info, redisFileName, "", "Redis Connected and Pinged")

		break
	}
}
