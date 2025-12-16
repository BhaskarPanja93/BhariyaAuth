package account

import (
	Config "BhariyaAuth/constants/config"
	ResponseModels "BhariyaAuth/models/responses"
	UserTypes "BhariyaAuth/models/users"
	Logger "BhariyaAuth/processors/logs"
	MailNotifier "BhariyaAuth/processors/mail"
	StringProcessor "BhariyaAuth/processors/string"
	Stores "BhariyaAuth/stores"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"golang.org/x/crypto/bcrypt"
)

func GetIDFromMail(mail string) (uint32, bool) {
	var userID uint32
	err := Stores.MySQLClient.QueryRow(`SELECT uid FROM users WHERE email = ? LIMIT 1`, mail).Scan(&userID)
	if err != nil {
		return 0, false
	}
	return userID, true
}

func GetMailFromID(userID uint32) (string, bool) {
	var mail string
	err := Stores.MySQLClient.QueryRow(`SELECT email FROM users WHERE uid = ? LIMIT 1`, userID).Scan(&mail)
	if err != nil {
		return "", false
	}
	return mail, true
}

func GetUserType(userID uint32) UserTypes.T {
	var userType string
	err := Stores.MySQLClient.QueryRow(`SELECT type FROM users WHERE uid = ? LIMIT 1`, userID).Scan(&userType)
	if err != nil {
		return UserTypes.All.Unknown
	}
	return UserTypes.Find(userType)
}

func BlacklistUser(userID uint32) bool {
	DeleteAllSessions(userID)
	_, err := Stores.MySQLClient.Exec("UPDATE users SET blocked = ? WHERE uid = ? LIMIT 1", true, userID)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[BlacklistUser] failed blocking [UID-%d] reason: %s", userID, err.Error()))
		return false
	}
	var mail string
	err = Stores.MySQLClient.QueryRow("SELECT email FROM users WHERE uid = ? LIMIT 1", userID).Scan(&mail)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[BlacklistUser] failed to fetch mail [UID-%d] reason: %s", userID, err.Error()))
		return false
	}
	MailNotifier.AccountBlacklisted(mail, 2)
	return true
}

func CheckUserIsBlacklisted(userID uint32) bool {
	var blocked bool
	err := Stores.MySQLClient.QueryRow("SELECT blocked FROM users WHERE uid = ? LIMIT 1", userID).Scan(&blocked)
	if err != nil {
		return true
	}
	return blocked
}

func DeleteAllSessions(userID uint32) bool {
	_, err := Stores.MySQLClient.Exec("DELETE FROM activities WHERE uid = ?", userID)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[DeleteAllSessions] failed [UID-%d] reason: %s", userID, err.Error()))
		return false
	}
	return true
}

func DeleteSession(userID uint32, refreshID uint16) {
	_, err := Stores.MySQLClient.Exec("DELETE FROM activities WHERE uid = ? AND refresh = ? LIMIT 1", userID, refreshID)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[DeleteSession] failed [UID-%d-RID-%d] reason: %s", userID, refreshID, err.Error()))
		DeleteSession(userID, refreshID)
		return
	}
}

func CheckSessionExists(userID uint32, refreshID uint16) bool {
	var blocked bool
	err := Stores.MySQLClient.QueryRow("SELECT EXISTS(SELECT 1 FROM activities WHERE uid = ? AND refresh = ?)", userID, refreshID).Scan(&blocked)
	if err != nil {
		return false
	}
	return blocked
}

func CheckUserHasPassword(userID uint32) bool {
	var hasPassword bool
	err := Stores.MySQLClient.QueryRow(`SELECT password IS NOT NULL AND password <> '' FROM users WHERE uid = ? LIMIT 1`, userID).Scan(&hasPassword)
	if err != nil {
		return false
	}
	return hasPassword
}

func UpdatePassword(userID uint32, password string) bool {
	if StringProcessor.PasswordIsStrong(password) {
		hash, ok := hashPassword(password)
		if !ok {
			return false
		}
		_, err := Stores.MySQLClient.Exec(`UPDATE users SET password = ? WHERE uid = ? LIMIT 1`, hash, userID)
		if err != nil {
			Logger.AccidentalFailure(fmt.Sprintf("[UpdatePassword] failed for [UID-%d] reason: %s", userID, err.Error()))
			return false
		}
		DeleteAllSessions(userID)
		return true
	}
	return false
}

func CheckPasswordMatches(userID uint32, password string) bool {
	if StringProcessor.PasswordIsStrong(password) {
		var hash string
		err := Stores.MySQLClient.QueryRow(`SELECT password FROM users WHERE uid = ? LIMIT 1`, userID).Scan(&hash)
		if err == nil {
			if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil {
				return true
			}
		} else {
			Logger.AccidentalFailure(fmt.Sprintf("[CheckPasswordMatches] failed for [UID-%d] reason: %s", userID, err.Error()))
		}
	}
	return false
}

func hashPassword(password string) (string, bool) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[HashPassword] error: %s", err.Error()))
		return "", false
	}
	return string(hash), true
}

func RecordNewUser(userID uint32, password string, mail string, name string) bool {
	var hash string
	var ok bool
	if password != "" {
		hash, ok = hashPassword(password)
		if !ok {
			return false
		}
	}
	now := time.Now().UTC()
	_, err := Stores.MySQLClient.Exec("INSERT INTO users VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, UserTypes.All.Viewer.Short, mail, name, false, hash, now)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[RecordNewUser] failed for [UID-%d-MAIL-%s] reason: %s", userID, mail, err.Error()))
		return false
	}
	MailNotifier.NewAccount(mail, 2)
	return true
}

func RecordReturningUser(mail string, IP string, ua string, refreshID uint16, userID uint32, rememberMe bool) bool {
	now := time.Now().UTC()
	var count int
	err := Stores.MySQLClient.QueryRow("SELECT COUNT(*) FROM activities WHERE uid = ?", userID).Scan(&count)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[RecordReturningUser] count query failed for [UID-%d]: %s", userID, err.Error()))
		return false
	}
	if count >= Config.MaxUserSessions {
		_, err = Stores.MySQLClient.Exec(`DELETE FROM activities WHERE uid = ? ORDER BY updated LIMIT 1`, userID)
		if err != nil {
			Logger.AccidentalFailure(fmt.Sprintf("[RecordReturningUser] failed deleting oldest session for [UID-%d]: %s", userID, err.Error()))
			return false
		}
	}
	_, err = Stores.MySQLClient.Exec(`INSERT INTO activities (uid, refresh, count, remembered, creation, updated, ua)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, userID, refreshID, 1, rememberMe, now, now, ua)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[RecordReturningUser] insert failed for [UID-%d-RID-%d]: %s", userID, refreshID, err.Error()))
		return false
	}
	MailNotifier.NewLogin(mail, IP, ua, 2)
	return true
}

func ServeAccountDetails() {
	channel := Stores.RedisClient.Subscribe(Stores.Ctx, Config.AccountDetailsRequestChannel).Channel()
	for message := range channel {
		var request ResponseModels.AccountDetailsRequestT
		if err := json.Unmarshal([]byte(message.Payload), &request); err != nil {
			Logger.AccidentalFailure(fmt.Sprintf("[ServeAccountDetails] Unmarshal failed %s reason: %s", message, err.Error()))
			continue
		}
		var response ResponseModels.AccountDetailsResponseT
		err := Stores.MySQLClient.QueryRow("SELECT uid, type, email, name, creation FROM users WHERE uid = ? LIMIT 1", request.UserID).
			Scan(&response.UserID, &response.Type, &response.Email, &response.Name, &response.Creation)
		if err != nil {
			Logger.AccidentalFailure(fmt.Sprintf("[ServeAccountDetails] Parse from DB failed for [UID-%d-SID-%s]: %s", request.UserID, request.ServerID, err.Error()))
			continue
		}
		Stores.RedisClient.Publish(Stores.Ctx, fmt.Sprintf("%s:%s", Config.AccountDetailsResponseChannel, request.ServerID), response)
	}
}
