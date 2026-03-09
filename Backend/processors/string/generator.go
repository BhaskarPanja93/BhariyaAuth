package string

import (
	"crypto/rand"
	"io"
)

const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	numbers = "0123456789"
)

func rejectionSample(dst []byte, charset string) {
	charsetLen := byte(len(charset))
	_max := byte(256 - (256 % int(charsetLen)))
	buf := make([]byte, len(dst))
	n := 0
	for n < len(dst) {
		if _, err := io.ReadFull(rand.Reader, buf); err != nil {
			panic(err)
		}
		for _, b := range buf {
			if b < _max {
				dst[n] = charset[b%charsetLen]
				n++
				if n == len(dst) {
					return
				}
			}
		}
	}
}

func SafeString(n uint16) string {
	if n == 0 {
		n = 1
	}
	b := make([]byte, n)
	rejectionSample(b, letters)
	return string(b)
}

func SafeNumber(n uint16) string {
	if n == 0 {
		n = 1
	}
	b := make([]byte, n)
	rejectionSample(b, numbers)
	return string(b)
}
