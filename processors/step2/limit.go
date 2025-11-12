package step2

import (
	Config "BhariyaAuth/constants/config"
	Stores "BhariyaAuth/stores"
	"fmt"
	"time"
)

func calculateResendDelay(value int64) time.Duration {
	return 10 * time.Second * time.Duration(value)
}

func calculateTTL(value int64) time.Duration {
	return time.Minute * time.Duration(value)
}

func craftKey(identifier string) string {
	return fmt.Sprintf("%s:%s", Config.RedisOTPRateLimit, identifier)
}

func CheckCanSendOTP(identifier string) (bool, int64, time.Duration) {
	key := craftKey(identifier)
	var timeRemaining time.Duration

	value, err := Stores.RedisClient.Get(Stores.Ctx, key).Int64()
	if err != nil || value == 0 {
		return true, 0, 0
	}

	ttl, err := Stores.RedisClient.TTL(Stores.Ctx, key).Result()
	if err != nil || ttl <= 0 {
		return true, value, 0
	}

	totalTTL := calculateTTL(value)
	elapsed := totalTTL - ttl
	resendDelay := calculateResendDelay(value)
	timeRemaining = resendDelay - elapsed

	canSend := timeRemaining <= 0
	if canSend {
		timeRemaining = 0
	}
	return canSend, value, timeRemaining
}

func RecordSendOTP(identifier string, value int64) time.Duration {
	key := craftKey(identifier)
	resendDuration := calculateResendDelay(value)
	Stores.RedisClient.Set(Stores.Ctx, key, value, calculateTTL(value))
	return resendDuration
}
