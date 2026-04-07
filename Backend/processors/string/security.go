package string

import (
	Secrets "BhariyaAuth/constants/secrets"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"unsafe"

	"github.com/goccy/go-json"
	"golang.org/x/crypto/bcrypt"
)

// Global encoding and cipher instances.
//
// b64:
// - Base64 URL-safe encoding without padding.
// - Strict mode ensures invalid input is rejected.
//
// aesGCM:
// - AEAD cipher (AES-GCM).
// - Provides authenticated encryption (confidentiality + integrity).
var b64 = base64.RawURLEncoding.Strict()
var aesGCM cipher.AEAD

// init initializes AES-GCM cipher using application secret key.
//
// Behavior:
// - Panics if cipher initialization fails (critical misconfiguration).
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

// BytesToB64 encodes binary data into URL-safe base64 string.
func BytesToB64(b []byte) string {
	return b64.EncodeToString(b)
}

// B64ToBytes decodes URL-safe base64 string into bytes.
//
// Returns error if input is malformed.
func B64ToBytes(s string) ([]byte, error) {
	return b64.DecodeString(s)
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
		return []byte{}, err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// EncryptToString encrypts data and encodes result as base64 string.
func EncryptToString(data []byte) (string, error) {
	ciphertext, err := EncryptToBytes(data)
	if err != nil {
		return "", err
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
		return nil, errors.New("b2b decrypt: data is too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
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
		return "", err
	}
	return BytesToB64(plaintext), nil
}

// EncryptInterfaceToBytes serializes and encrypts arbitrary struct.
func EncryptInterfaceToBytes(v interface{}) ([]byte, error) {
	marshalled, err := json.Marshal(v)
	if err != nil {
		return []byte{}, err
	}
	return EncryptToBytes(marshalled)
}

// EncryptInterfaceToString serializes, encrypts, and encodes to base64.
func EncryptInterfaceToString(v interface{}) (string, error) {
	ciphertext, err := EncryptInterfaceToBytes(v)
	if err != nil {
		return "", err
	}
	return BytesToB64(ciphertext), nil
}

// DecryptInterfaceFromBytes decrypts and deserializes into struct.
func DecryptInterfaceFromBytes(data []byte, v interface{}) error {
	plaintext, err := DecryptToBytes(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(plaintext, v)
}

// DecryptInterfaceFromString decodes, decrypts, and deserializes.
func DecryptInterfaceFromString(data string, v interface{}) error {
	ciphertext, err := B64ToBytes(data)
	if err != nil {
		return err
	}
	return DecryptInterfaceFromBytes(ciphertext, v)
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
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// EncryptUserID encrypts int32 using unsafe pointer conversion.
func EncryptUserID(userID int32) (string, error) {
	return EncryptToString((*(*[4]byte)(unsafe.Pointer(&userID)))[:])
}

// EncryptDeviceID encrypts int16 using unsafe pointer conversion.
func EncryptDeviceID(deviceID int16) (string, error) {
	return EncryptToString((*(*[2]byte)(unsafe.Pointer(&deviceID)))[:])
}

// DecryptUserID decrypts and converts bytes back to int32.
//
// WARNING:
// - Assumes byte slice has correct length and alignment.
func DecryptUserID(data string) (int32, error) {
	ciphertext, err := B64ToBytes(data)
	if err != nil {
		return 0, err
	}
	bytes, err := DecryptToBytes(ciphertext)
	if err != nil {
		return 0, err
	}
	return *(*int32)(unsafe.Pointer(&bytes[0])), nil
}

// DecryptDeviceID decrypts and converts bytes back to int16.
//
// WARNING:
// - Assumes byte slice has correct length and alignment.
func DecryptDeviceID(data string) (int16, error) {
	ciphertext, err := B64ToBytes(data)
	if err != nil {
		return 0, err
	}
	bytes, err := DecryptToBytes(ciphertext)
	if err != nil {
		return 0, err
	}
	return *(*int16)(unsafe.Pointer(&bytes[0])), nil
}
