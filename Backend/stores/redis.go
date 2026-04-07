package stores

import (
	Config "BhariyaAuth/constants/config"
	"fmt"
	"time"

	Secrets "BhariyaAuth/constants/secrets"

	"github.com/redis/go-redis/v9"
)

// RedisClient is a global Redis client instance.
//
// Notes:
// - Initialized once via ConnectRedis().
// - Shared across application.
// - Thread-safe (go-redis client is concurrency-safe).
var RedisClient *redis.Client

// ConnectRedis initializes a Redis connection with automatic fallback and retry.
//
// Overview:
// This function attempts to establish a connection to Redis using:
//  1. UNIX socket (preferred for local deployments).
//  2. TCP/IP fallback (for remote or containerized environments).
//
// Flow:
//
//	try socket → ping → on failure switch to TCP → retry → loop until success
//
// Behavior:
// - Retries indefinitely until connection succeeds.
// - Alternates between socket and TCP on failure.
// - Waits 2 seconds between retries.
//
// Configuration:
// - Uses Secrets for connection details.
// - Uses connection pooling for performance.
//
// Returns:
// - No return value (blocks until successful connection).
//
// Important:
// - This function blocks indefinitely until Redis is reachable.
func ConnectRedis() {

	// Prevent re-initialization if already connected
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

			// Connection failed → cleanup and retry
			fmt.Println("Redis connection failed:", err.Error())

			_ = RedisClient.Close()

			// Toggle connection method
			useSocket = !useSocket

			// Backoff before retry
			time.Sleep(2 * time.Second)

			continue
		}

		// Successful connection
		break
	}

	fmt.Println("Redis connection established successfully")
}
