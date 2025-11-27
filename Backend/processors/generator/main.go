package generator

import (
	Logger "BhariyaAuth/processors/logs"
	CryptoRand "crypto/rand"
	"fmt"
	"math/big"
)

const _letters = "_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const _numbers = "0123456789"

func SafeString(nBytes uint16) string {
	if nBytes <= 0 {
		fmt.Println("nBytes must be greater than 0")
		return SafeString(1)
	}
	b := make([]byte, nBytes)
	for i := range b {
		num, err := CryptoRand.Int(CryptoRand.Reader, big.NewInt(int64(len(_letters))))
		if err != nil {
			Logger.AccidentalFailure(fmt.Sprintf("[SafeString] failed for length [%d] error: %s", nBytes, err.Error()))
			return SafeString(nBytes)
		}
		b[i] = _letters[num.Int64()]
	}
	return string(b)
}

func SafeNumber(nBytes uint16) string {
	if nBytes <= 0 {
		fmt.Println("nBytes must be greater than 0")
		return SafeNumber(1)
	}
	b := make([]byte, nBytes)
	for i := range b {
		num, err := CryptoRand.Int(CryptoRand.Reader, big.NewInt(int64(len(_numbers))))
		if err != nil {
			Logger.AccidentalFailure(fmt.Sprintf("[SafeNumber] failed for length [%d] error: %s", nBytes, err.Error()))
			return SafeNumber(nBytes)
		}
		b[i] = _numbers[num.Int64()]
	}
	return string(b)
}

func UserID() uint32 {
	b := make([]byte, 3)
	if _, err := CryptoRand.Read(b); err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("UserID failed: %s", err.Error()))
		return UserID()
	}
	var val uint32
	var i uint16 = 0
	for i = 0; i < 3; i++ {
		shift := uint((2 - i) * 8)
		val |= uint32(b[i]) << shift
	}
	if val == 0 {
		return UserID()
	}
	return val
}

func RefreshID() uint16 {
	b := make([]byte, 2)
	if _, err := CryptoRand.Read(b); err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("RefreshID failed: %s", err.Error()))
		return RefreshID()
	}
	var val uint16
	var i uint16 = 0
	for i = 0; i < 2; i++ {
		shift := uint((1 - i) * 8)
		val |= uint16(b[i]) << shift
	}
	if val == 0 {
		return RefreshID()
	}
	return val
}
