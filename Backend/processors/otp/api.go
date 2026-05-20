package otp

import (
	MailModels "BhariyaAuth/models/mails"
	MailNotifier "BhariyaAuth/processors/mail"
	StringProcessor "BhariyaAuth/processors/string"
	"errors"
	"time"
)

func Send(address string, model MailModels.T, identifier string) (string, time.Duration, error) {
	key := address + ":" + identifier

	canSend, count, wait := checkCanSend(key)
	if !canSend {
		return "", wait, errors.New("otp resend rate limited")
	}

	otpValue := StringProcessor.SafeNumber(6)
	verificationToken := StringProcessor.SafeString(12)

	if err := MailNotifier.OTP(address, otpValue, model, 2); err != nil {
		return "", wait, errors.New("otp send failed: " + err.Error())
	}

	delay := recordSend(key, verificationToken, otpValue, count)
	return verificationToken, delay, nil
}

func Validate(verification, otp string) bool {
	val, ok := otpStore.Load(verification)
	if !ok {
		return false
	}

	entry := val.(*otpEntry)
	if time.Now().After(entry.expires) || otp != entry.otp {
		return false
	}

	otpStore.Delete(verification)
	return true
}
