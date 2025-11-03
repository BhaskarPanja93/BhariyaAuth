package generator

import (
	CryptoRand "crypto/rand"
	"fmt"
	"math/big"
	MathRand "math/rand"
)

const _letters = "_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func UnsafeString(nBytes uint16) string {
	if nBytes <= 0 {
		fmt.Println("nBytes must be greater than 0")
		return UnsafeString(1)
	}
	b := make([]byte, nBytes)
	for i := range b {
		b[i] = _letters[MathRand.Intn(len(_letters))]
	}
	return string(b)
}

func SafeString(nBytes uint16) string {
	if nBytes <= 0 {
		fmt.Println("nBytes must be greater than 0")
		return SafeString(1)
	}
	b := make([]byte, nBytes)
	for i := range b {
		num, err := CryptoRand.Int(CryptoRand.Reader, big.NewInt(int64(len(_letters))))
		if err != nil {
			return SafeString(nBytes)
		}
		b[i] = _letters[num.Int64()]
	}
	return string(b)
}

func UserID() uint32 {
	b := make([]byte, 3)
	if _, err := CryptoRand.Read(b); err != nil {
		return UserID()
	}
	var val uint32
	var i uint16 = 0
	for i = 0; i < 3; i++ {
		shift := uint((2 - i) * 8)
		val |= uint32(b[i]) << shift
	}
	return val
}

func RefreshID() uint16 {
	b := make([]byte, 2)
	if _, err := CryptoRand.Read(b); err != nil {
		return RefreshID()
	}
	var val uint16
	var i uint16 = 0
	for i = 0; i < 2; i++ {
		shift := uint((1 - i) * 8)
		val |= uint16(b[i]) << shift
	}
	return val
}
