package account

import (
	Config "BhariyaAuth/constants/config"
	MailModels "BhariyaAuth/models/mails"
	MailNotifier "BhariyaAuth/processors/mail"
	StringProcessor "BhariyaAuth/processors/string"
	Stores "BhariyaAuth/stores"
	"errors"

	"github.com/gofiber/fiber/v3"
)

func RecordNewUser(ctx fiber.Ctx, userType string, password string, mail string, name string) (int32, error) {
	IP := ctx.IP()
	ua := ctx.Get("User-Agent")
	var userID int32
	var hashBytes []byte
	var hash string
	var err error
	if password != "" {
		hashBytes, err = StringProcessor.HashPassword(password)
		if err != nil {
			return userID, errors.New("Record new user: " + err.Error())
		}
		hash = string(hashBytes)
	}
	err = Stores.SQLClient.QueryRow(Config.CtxBG,
		`INSERT INTO users (type, mail, name, blocked, pw_hash) VALUES ($1, $2, $3, $4, $5) RETURNING user_id`,
		userType, mail, name, false, hash).
		Scan(&userID)
	if err != nil {
		return 0, errors.New("Record new user - SQL query: " + err.Error())
	}
	os, device, browser := StringProcessor.ParseUA(ua)
	MailNotifier.SignUp(mail, name, MailModels.SignUpComplete, IP, os, device, browser, 2)
	return userID, nil
}

func RecordReturningUser(ctx fiber.Ctx, mail string, userID int32, rememberMe bool, sendMail bool) (int16, error) {
	IP := ctx.IP()
	ua := ctx.Get("User-Agent")
	var deviceID int16
	os, device, browser := StringProcessor.ParseUA(ua)
	err := Stores.SQLClient.QueryRow(Config.CtxBG,
		`INSERT INTO devices (user_id, remembered, os, device, browser) VALUES ($1, $2, $3, $4, $5) RETURNING device_id`,
		userID, rememberMe, os, device, browser).
		Scan(&deviceID)
	if err != nil {
		return deviceID, errors.New("Record returning user - SQL query: " + err.Error())
	}
	if sendMail {
		MailNotifier.SignIn(mail, MailModels.SignInComplete, IP, os, device, browser, 2)
	}
	return deviceID, nil
}
