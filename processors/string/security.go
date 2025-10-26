package string

import (
	Secrets "BhariyaAuth/constants/secrets"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

var aesGCM cipher.AEAD

func init() {
	block, err := aes.NewCipher(Secrets.AESGCMKey)
	if err != nil {
		panic(err)
	}
	aesGCM, err = cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
}

func Encrypt(data []byte) (string, bool) {
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", false
	}
	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), true
}

func Decrypt(token string) ([]byte, bool) {
	ciphertext, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, false
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, false
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, false
	}
	return plaintext, true
}
