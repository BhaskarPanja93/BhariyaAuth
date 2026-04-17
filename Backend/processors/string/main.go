package string

import (
	Secrets "BhariyaAuth/constants/secrets"
	Logs "BhariyaAuth/processors/logs"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"regexp"

	"github.com/medama-io/go-useragent"
)

// Precompiled regex for email validation.
//
// Pattern:
// - Local part: alphanumeric + ._%+-
// - Domain: alphanumeric + dots/hyphens
// - TLD: minimum 2 characters
//
// Note:
// - This is a pragmatic (not RFC-complete) validation.
// - Designed for typical web use, not full RFC 5322 compliance.
var emailRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[a-zA-Z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+)*|"(?:[\x20-\x7E]|\\[\x20-\x7E])*")@(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// uaParser is a reusable User-Agent parser instance.
//
// Used to extract:
// - OS
// - Device
// - Browser + version
var uaParser = useragent.NewParser()

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
		Logs.RootLogger.Add(Logs.Error, "processors/string/security", "", "New cipher failed: "+err.Error())
		panic(err)
	}

	aesGCM, err = cipher.NewGCM(block)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, "processors/string/security", "", "New GCM failed: "+err.Error())
		panic(err)
	}
}
