package string

import (
	Secrets "BhariyaAuth/constants/secrets"
	Logger "BhariyaAuth/processors/logs"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

var aesGCM cipher.AEAD

func init() {
	block, err := aes.NewCipher(Secrets.AESGCMKey)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("NewCipher failed: %s", err.Error()))
		panic(err)
	}
	aesGCM, err = cipher.NewGCM(block)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("NewGCM failed: %s", err.Error()))
		panic(err)
	}
}

func Encrypt(data []byte) (string, bool) {
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[Encrypt] error: %s", err.Error()))
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
		Logger.AccidentalFailure(fmt.Sprintf("[Decrypt] error: %s", err.Error()))
		return nil, false
	}
	return plaintext, true
}
