package tokens

import (
	UsersModel "BhariyaAuth/models/users"
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

type SSOStateT struct {
	Provider      string    `json:"pro"`
	Expiry        time.Time `json:"exp"`
	FrontendState string    `json:"fro"`
	Origin        string    `json:"ori"`
	RememberMe    bool      `json:"rem"`
}

type AccessTokenT struct {
	UserID         uint32       `json:"uid"`
	RefreshID      uint16       `json:"rid"`
	RefreshIndex   uint16       `json:"rin"`
	UserType       UsersModel.T `json:"typ"`
	AccessCreated  time.Time    `json:"aat"`
	RefreshCreated time.Time    `json:"lat"`
	RememberMe     bool         `json:"rem"`
}

type RefreshTokenT struct {
	UserID         uint32       `json:"uid"`
	RefreshID      uint16       `json:"rid"`
	RefreshIndex   uint16       `json:"rin"`
	RefreshCreated time.Time    `json:"lat"`
	RefreshUpdated time.Time    `json:"rat"`
	UserType       UsersModel.T `json:"typ"`
	CSRF           string       `json:"csr"`
	RememberMe     bool         `json:"rem"`
	IdentifierType string       `json:"siu"`
}

type NewTokenT struct {
	AccessToken  string
	RefreshToken string
	CSRF         string
	RefreshID    uint16
	RefreshIndex uint16
	RememberMe   bool
}
