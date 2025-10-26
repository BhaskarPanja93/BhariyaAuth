package string

import (
	CryptoRand "crypto/rand"
	"fmt"
	"math/big"
	MathRand "math/rand"
)

const _letters = "_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateUnsafeString(nBytes uint16) string {
	if nBytes <= 0 {
		fmt.Println("nBytes must be greater than 0")
		return GenerateUnsafeString(1)
	}
	b := make([]byte, nBytes)
	for i := range b {
		b[i] = _letters[MathRand.Intn(len(_letters))]
	}
	return string(b)
}

func GenerateSafeString(nBytes uint16) string {
	if nBytes <= 0 {
		fmt.Println("nBytes must be greater than 0")
		return GenerateSafeString(1)
	}
	b := make([]byte, nBytes)
	for i := range b {
		num, err := CryptoRand.Int(CryptoRand.Reader, big.NewInt(int64(len(_letters))))
		if err != nil {
			return GenerateSafeString(nBytes)
		}
		b[i] = _letters[num.Int64()]
	}
	return string(b)
}

func GenerateUserID() uint32 {
	b := make([]byte, 3)
	if _, err := CryptoRand.Read(b); err != nil {
		return GenerateUserID()
	}
	var val uint32
	var i uint16 = 0
	for i = 0; i < 3; i++ {
		shift := uint((2 - i) * 8)
		val |= uint32(b[i]) << shift
	}
	return val
}

func GenerateRefreshID() uint16 {
	b := make([]byte, 2)
	if _, err := CryptoRand.Read(b); err != nil {
		return GenerateRefreshID()
	}
	var val uint16
	var i uint16 = 0
	for i = 0; i < 2; i++ {
		shift := uint((1 - i) * 8)
		val |= uint16(b[i]) << shift
	}
	return val
}

func GenerateSafeInt64(nBytes uint16) uint64 {
	if nBytes <= 0 || nBytes > 8 {
		fmt.Println("nBytes must be between 1 and 8")
		return GenerateSafeInt64(8)
	}
	b := make([]byte, nBytes)
	if _, err := CryptoRand.Read(b); err != nil {
		return GenerateSafeInt64(nBytes)
	}
	var val uint64
	var i uint16 = 0
	for i = 0; i < nBytes; i++ {
		shift := uint((nBytes - i - 1) * 8)
		val |= uint64(b[i]) << shift
	}
	return val
}
