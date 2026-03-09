package tokens

import (
	UsersTypes "BhariyaAuth/models/users"
	"time"
)

type SignInT struct {
	TokenType    string `json:"tt"`
	User         int32  `json:"uid"`
	RememberMe   bool   `json:"rem"`
	Step2Process string `json:"2t"`
	Step2Code    string `json:"2c"`
	Mail         string `json:"mail"`
}

type SignUpT struct {
	TokenType  string `json:"tt"`
	Mail       string `json:"mail"`
	RememberMe bool   `json:"rem"`
	Name       string `json:"name"`
	Password   string `json:"pass"`
	Step2Code  string `json:"2c"`
}

type PasswordResetT struct {
	TokenType string `json:"tt"`
	Mail      string `json:"mail"`
	User      int32  `json:"uid"`
	Step2Code string `json:"2c"`
}

type SSOStateT struct {
	Provider   string    `json:"pro"`
	Expiry     time.Time `json:"exp"`
	RememberMe bool      `json:"rem"`
}

type MFATokenT struct {
	TokenType string    `json:"tt"`
	Step2Code string    `json:"2c"`
	User      int32     `json:"uid"`
	Created   time.Time `json:"cre"`
	Verified  bool      `json:"ver"`
}

type AccessTokenT struct {
	User         int32        `json:"uid"`
	Refresh      int16        `json:"rid"`
	UserType     UsersTypes.T `json:"typ"`
	AccessExpiry time.Time    `json:"axe"`
	RememberMe   bool         `json:"rem"`
}

type RefreshTokenT struct {
	User           int32        `json:"uid"`
	Refresh        int16        `json:"rid"`
	Visits         int16        `json:"vis"`
	Created        time.Time    `json:"cre"`
	Updated        time.Time    `json:"upd"`
	Expiry         time.Time    `json:"exp"`
	UserType       UsersTypes.T `json:"typ"`
	CSRF           string       `json:"csrf"`
	RememberMe     bool         `json:"rem"`
	IdentifierType string       `json:"it"`
}

type NewTokenCombinedT struct {
	AccessToken   string
	RefreshToken  string
	AccessExpires time.Time
	CSRF          string
	RememberMe    bool
}
