package tokens

import (
	UsersTypes "BhariyaAuth/models/users"
	"time"
)

type SignInT struct {
	TokenType    string `json:"tt"`
	UserID       uint32 `json:"uid"`
	RememberMe   bool   `json:"remember"`
	Step2Process string `json:"2_type"`
	Step2Code    string `json:"2_code"`
	Mail         string `json:"mail"`
}

type SignUpT struct {
	TokenType  string `json:"tt"`
	Mail       string `json:"mail"`
	RememberMe bool   `json:"remember"`
	Name       string `json:"name"`
	Password   string `json:"password"`
	Step2Code  string `json:"2_code"`
}

type PasswordResetT struct {
	TokenType string `json:"tt"`
	UserID    uint32 `json:"uid"`
	Step2Code string `json:"2_code"`
}

type SSOStateT struct {
	Provider   string    `json:"pro"`
	Expiry     time.Time `json:"exp"`
	RememberMe bool      `json:"rem"`
}

type MFATokenT struct {
	TokenType string    `json:"tt"`
	Step2Code string    `json:"2_code"`
	UserID    uint32    `json:"uid"`
	Creation  time.Time `json:"cre"`
	Verified  bool      `json:"ver"`
}

type AccessTokenT struct {
	UserID       uint32       `json:"uid"`
	RefreshID    uint16       `json:"rid"`
	UserType     UsersTypes.T `json:"typ"`
	AccessExpiry time.Time    `json:"axe"`
	RememberMe   bool         `json:"rem"`
}

type RefreshTokenT struct {
	UserID         uint32       `json:"uid"`
	RefreshID      uint16       `json:"rid"`
	RefreshIndex   uint16       `json:"rin"`
	RefreshCreated time.Time    `json:"rca"`
	RefreshUpdated time.Time    `json:"rua"`
	RefreshExpiry  time.Time    `json:"rxa"`
	UserType       UsersTypes.T `json:"typ"`
	CSRF           string       `json:"csr"`
	RememberMe     bool         `json:"rem"`
	IdentifierType string       `json:"siu"`
}

type NewTokenCombinedT struct {
	AccessToken  string
	RefreshToken string
	CSRF         string
	RememberMe   bool
}
