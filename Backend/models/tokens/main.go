package tokens

import (
	"time"
)

// All tokens  that are either generated, for eg, Step1 of a process, Cookie etc

type SignIn struct {
	TokenType    string `json:"tt"`
	UserID       int32  `json:"uid"`
	Remember     bool   `json:"rem"`
	Step2Process string `json:"2t"`
	Step2Code    string `json:"2c"`
	MailAddress  string `json:"add"`
}

type SignUp struct {
	TokenType   string `json:"tt"`
	MailAddress string `json:"add"`
	Remember    bool   `json:"rem"`
	Name        string `json:"name"`
	Password    string `json:"pass"`
	Step2Code   string `json:"2c"`
}

type PasswordReset struct {
	TokenType   string `json:"tt"`
	MailAddress string `json:"add"`
	UserID      int32  `json:"uid"`
	Step2Code   string `json:"2c"`
}

type SSOState struct {
	TokenType string    `json:"tt"`
	Provider  string    `json:"pro"`
	State     string    `json:"st"`
	Expiry    time.Time `json:"exp"`
	Remember  bool      `json:"rem"`
}

type MFAToken struct {
	TokenType string    `json:"tt"`
	Step2Code string    `json:"2c"`
	UserID    int32     `json:"uid"`
	DeviceID  int16     `json:"did"`
	Created   time.Time `json:"cre"`
	Verified  bool      `json:"ver"`
}

type AccessToken struct {
	TokenType string    `json:"tt"`
	UserID    int32     `json:"uid"`
	DeviceID  int16     `json:"did"`
	UserType  string    `json:"typ"`
	Expiry    time.Time `json:"exp"`
	Remember  bool      `json:"rem"`
}

type RefreshToken struct {
	TokenType      string    `json:"tt"`
	UserID         int32     `json:"uid"`
	DeviceID       int16     `json:"did"`
	Visits         int16     `json:"vis"`
	Created        time.Time `json:"cre"`
	Updated        time.Time `json:"upd"`
	Expiry         time.Time `json:"exp"`
	UserType       string    `json:"typ"`
	CSRF           string    `json:"csrf"`
	Remember       bool      `json:"rem"`
	IdentifierType string    `json:"it"`
}

type NewTokenCombined struct {
	AccessToken   string
	RefreshToken  string
	AccessExpires time.Time
	CSRF          string
	RememberMe    bool
}
