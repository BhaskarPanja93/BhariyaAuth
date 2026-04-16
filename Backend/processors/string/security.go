package string

import (
	"crypto/rand"
	"errors"
	"io"

	"github.com/bytedance/sonic"
	"golang.org/x/crypto/bcrypt"
)

// BytesToB64 encodes binary data into URL-safe base64 string.
func BytesToB64(b []byte) string {
	return b64.EncodeToString(b)
}

// B64ToBytes decodes URL-safe base64 string into bytes.
//
// Returns error if input is malformed.
func B64ToBytes(s string) ([]byte, error) {
	data, err := b64.DecodeString(s)
	if err != nil {
		return nil, errors.New("b64 decode failed: " + err.Error())
	}
	return data, nil
}

// EncryptToBytes encrypts plaintext using AES-GCM.
//
// Flow:
//
//	generate nonce → seal data → prepend nonce
//
// Output format:
//
//	[nonce | ciphertext]
//
// Security:
// - Nonce must be unique per encryption (guaranteed via crypto/rand).
// - AES-GCM ensures both encryption and authentication.
func EncryptToBytes(data []byte) ([]byte, error) {
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.New("encrypt to bytes failed - Read Full: " + err.Error())
	}
	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// EncryptToString encrypts data and encodes result as base64 string.
func EncryptToString(data []byte) (string, error) {
	ciphertext, err := EncryptToBytes(data)
	if err != nil {
		return "", errors.New("encrypt to string failed: " + err.Error())
	}
	return BytesToB64(ciphertext), nil
}

// DecryptToBytes decrypts AES-GCM ciphertext.
//
// Input format:
//
//	[nonce | ciphertext]
//
// Returns:
// - plaintext if authentication succeeds.
// - error if:
//   - data is too short
//   - authentication fails
func DecryptToBytes(data []byte) ([]byte, error) {
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("decrypt to bytes failed: data is too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decrypt to bytes failed - Open: " + err.Error())
	}
	return plaintext, nil
}

// DecryptToString decrypts data and returns base64-encoded plaintext.
//
// NOTE:
// - This returns BASE64(plaintext), not raw string.
func DecryptToString(data []byte) (string, error) {
	plaintext, err := DecryptToBytes(data)
	if err != nil {
		return "", errors.New("decrypt to string: " + err.Error())
	}
	return BytesToB64(plaintext), nil
}

// EncryptInterfaceToBytes serializes and encrypts arbitrary struct.
func EncryptInterfaceToBytes(v interface{}) ([]byte, error) {
	marshalled, err := sonic.Marshal(v)
	if err != nil {
		return nil, errors.New("encrypt interface to bytes - Marshal: " + err.Error())
	}
	data, err := EncryptToBytes(marshalled)
	if err != nil {
		return nil, errors.New("encrypt interface to bytes: " + err.Error())
	}
	return data, nil
}

// DecryptInterfaceFromBytes decrypts and deserializes into struct.
func DecryptInterfaceFromBytes(data []byte, v interface{}) error {
	plaintext, err := DecryptToBytes(data)
	if err != nil {
		return errors.New("decrypt interface from bytes: " + err.Error())
	}
	err = sonic.Unmarshal(plaintext, v)
	if err != nil {
		return errors.New("decrypt interface from bytes - Unmarshal: " + err.Error())
	}
	return nil
}

// EncryptInterfaceToB64 serializes, encrypts, and encodes to base64.
func EncryptInterfaceToB64(v interface{}) (string, error) {
	ciphertext, err := EncryptInterfaceToBytes(v)
	if err != nil {
		return "", errors.New("encrypt interface to string: " + err.Error())
	}
	return BytesToB64(ciphertext), nil
}

// DecryptInterfaceFromB64 decodes, decrypts, and deserializes.
func DecryptInterfaceFromB64(data string, v interface{}) error {
	ciphertext, err := B64ToBytes(data)
	if err != nil {
		return errors.New("decrypt interface from string: " + err.Error())
	}
	err = DecryptInterfaceFromBytes(ciphertext, v)
	if err != nil {
		return errors.New("decrypt interface from string: " + err.Error())
	}
	return nil
}

// HashPassword hashes password using bcrypt.
//
// Notes:
// - Uses DefaultCost (secure baseline).
// - Output includes salt + hash.
//
// Returns:
// - hashed password bytes.
// - error if hashing fails.
func HashPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("hash password - Generate from password: " + err.Error())
	}
	return hash, nil
}
