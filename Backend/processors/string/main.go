package string

import (
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
