package otp

import (
	Config "BhariyaAuth/constants/config"
	Generators "BhariyaAuth/processors/generator"
	MailNotifier "BhariyaAuth/processors/mail"
	Stores "BhariyaAuth/stores"
	"fmt"
	"sync"
	"time"
)

type otpEntry struct {
	Value     int64
	CreatedAt time.Time
}

var (
	otpStore = struct {
		sync.Mutex
		data map[string]otpEntry
	}{
		data: make(map[string]otpEntry),
	}
)

func init() {
	go func() {
		for {
			time.Sleep(time.Minute * 10)
			now := time.Now()

			otpStore.Lock()
			for k, v := range otpStore.data {
				if now.Sub(v.CreatedAt) >= calculateTTL(v.Value) {
					delete(otpStore.data, k)
				}
			}
			otpStore.Unlock()
		}
	}()
}

func calculateResendDelay(value int64) time.Duration {
	return 10 * time.Second * time.Duration(value)
}

func calculateTTL(value int64) time.Duration {
	return time.Minute * time.Duration(value)
}

func checkCanSend(identifier string) (bool, int64, time.Duration) {
	otpStore.Lock()
	entry, exists := otpStore.data[identifier]
	otpStore.Unlock()
	if !exists {
		return true, 0, 0
	}
	value := entry.Value
	totalTTL := calculateTTL(value)
	elapsed := time.Since(entry.CreatedAt)
	if elapsed >= totalTTL {
		otpStore.Lock()
		delete(otpStore.data, identifier)
		otpStore.Unlock()
		return true, value, 0
	}
	resendDelay := calculateResendDelay(value)
	timeRemaining := resendDelay - elapsed
	canSend := timeRemaining <= 0
	if canSend {
		timeRemaining = 0
	}
	return canSend, value, timeRemaining
}

func recordSent(identifier string, value int64) time.Duration {
	now := time.Now()
	otpStore.Lock()
	otpStore.data[identifier] = otpEntry{
		Value:     value,
		CreatedAt: now,
	}
	otpStore.Unlock()
	return calculateResendDelay(value)
}

func Send(mail string, subject string, header string, ignorable bool, identifier string) (string, time.Duration) {
	rateLimitKey := fmt.Sprintf("%s:%s", mail, identifier)
	canSend, alreadySentCount, currentDelay := checkCanSend(rateLimitKey)
	if canSend {
		otp := Generators.SafeNumber(6)
		if success := MailNotifier.OTP(mail, otp, subject, header, ignorable, 2); !success {
			return "", currentDelay
		}
		verification := Generators.SafeString(10)
		key := fmt.Sprintf("%s:%s", Config.RedisServerOTPVerification, verification)
		Stores.RedisClient.Set(Stores.Ctx, key, otp, 5*time.Minute)
		currentDelay = recordSent(rateLimitKey, alreadySentCount+1)
		return verification, currentDelay
	}
	return "", currentDelay
}

func Validate(verification, otp string) bool {
	key := fmt.Sprintf("%s:%s", Config.RedisServerOTPVerification, verification)
	value, _ := Stores.RedisClient.Get(Stores.Ctx, key).Result()
	if value == otp && otp != "" {
		Stores.RedisClient.Del(Stores.Ctx, key)
		return true
	}
	return false
}
