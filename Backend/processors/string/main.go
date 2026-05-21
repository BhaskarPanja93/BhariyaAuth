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

var emailRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[a-zA-Z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+)*|"(?:[\x20-\x7E]|\\[\x20-\x7E])*")@(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

var uaParser = useragent.NewParser()

var b64 = base64.RawURLEncoding.Strict()
var aesGCM cipher.AEAD

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
