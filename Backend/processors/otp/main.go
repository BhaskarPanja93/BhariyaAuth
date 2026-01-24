package otp

import (
	Generators "BhariyaAuth/processors/generator"
	MailNotifier "BhariyaAuth/processors/mail"
	"fmt"
	"sync"
	"time"
)

type LimiterEntryT struct {
	SentCount uint16
	LastSent  time.Time
}

type VerificationEntryT struct {
	OTP     string
	Expires time.Time
}

var (
	limiterMap = struct {
		sync.Mutex
		data map[string]LimiterEntryT
	}{
		data: make(map[string]LimiterEntryT),
	}

	verificationMap = struct {
		sync.Mutex
		data map[string]VerificationEntryT
	}{
		data: make(map[string]VerificationEntryT),
	}
)

func init() {
	go func() {
		for {
			time.Sleep(time.Minute * 10)
			now := time.Now()
			limiterMap.Lock()
			for k, v := range limiterMap.data {
				if now.Sub(v.LastSent) >= calculateLimiterEntryTTL(v.SentCount) {
					delete(limiterMap.data, k)
				}
			}
			limiterMap.Unlock()
			verificationMap.Lock()
			for k, v := range verificationMap.data {
				if now.After(v.Expires) {
					delete(verificationMap.data, k)
				}
			}
			verificationMap.Unlock()
		}
	}()
}

func calculateResendDelay(sentCount uint16) time.Duration {
	return 10 * time.Second * time.Duration(sentCount)
}

func calculateLimiterEntryTTL(sentCount uint16) time.Duration {
	return time.Minute * time.Duration(sentCount)
}

func checkCanSend(identifier string) (bool, uint16, time.Duration) {
	limiterMap.Lock()
	entry, exists := limiterMap.data[identifier]
	limiterMap.Unlock()
	if !exists {
		return true, 0, 0
	}
	value := entry.SentCount
	totalTTL := calculateLimiterEntryTTL(value)
	elapsed := time.Since(entry.LastSent)
	if elapsed >= totalTTL {
		limiterMap.Lock()
		delete(limiterMap.data, identifier)
		limiterMap.Unlock()
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

func recordSent(identifier string, sentCount uint16, verification, OTP string) time.Duration {
	now := time.Now()
	limiterMap.Lock()
	limiterMap.data[identifier] = LimiterEntryT{
		SentCount: sentCount,
		LastSent:  now,
	}
	limiterMap.Unlock()
	verificationMap.Lock()
	verificationMap.data[verification] = VerificationEntryT{
		OTP:     OTP,
		Expires: now.Add(5 * time.Minute),
	}
	verificationMap.Unlock()
	return calculateResendDelay(sentCount)
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
		currentDelay = recordSent(rateLimitKey, alreadySentCount+1, verification, otp)
		return verification, currentDelay
	}
	return "", currentDelay
}

func Validate(verification, otp string) bool {
	now := time.Now().UTC()
	verificationMap.Lock()
	entry, exists := verificationMap.data[verification]
	verificationMap.Unlock()
	if !exists {
		return false
	}
	if otp == "" || entry.OTP != otp || entry.Expires.Before(now) {
		return false
	}
	verificationMap.Lock()
	delete(verificationMap.data, verification)
	verificationMap.Unlock()
	return true
}
