package account

import (
	Config "BhariyaAuth/constants/config"
	TokenModels "BhariyaAuth/models/tokens"
	UserTypes "BhariyaAuth/models/users"
	Logger "BhariyaAuth/processors/logs"
	Stores "BhariyaAuth/stores"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"golang.org/x/crypto/bcrypt"
)

func IDExists(userID uint32) bool {
	var exists bool
	err := Stores.MySQLClient.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE uid = ?)`, userID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func GetIDFromMail(mail string) (uint32, bool) {
	var userID uint32
	err := Stores.MySQLClient.QueryRow(`SELECT uid FROM users WHERE email = ? LIMIT 1`, mail).Scan(&userID)
	if err != nil {
		return 0, false
	}
	return userID, true
}

func GetUserType(userID uint32) string {
	var userType string
	err := Stores.MySQLClient.QueryRow(`SELECT type FROM users WHERE uid = ?`, userID).Scan(&userType)
	if err != nil {
		return ""
	}
	return userType
}

func BlacklistUser(userID uint32) {
	CloseAllSessions(userID)
	_, err := Stores.MySQLClient.Exec("UPDATE users SET blocked = ? WHERE uid = ?", true, userID)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Blacklisting user failed: %s", err.Error()))
		BlacklistUser(userID)
		return
	}
	rows, err := Stores.MySQLClient.Query("SELECT refresh FROM activities WHERE uid = ? AND blocked = ?", userID, false)
	if err != nil {
		BlacklistUser(userID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var refreshID uint16
		if err = rows.Scan(&refreshID); err != nil {
			return
		}
	}
	_, err = Stores.MySQLClient.Exec("UPDATE activities SET blocked = ? WHERE uid = ? AND blocked = ?", true, userID, false)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Blacklisting refresh failed: %s", err.Error()))
		return
	}
}

func UserIsBlacklisted(userID uint32) bool {
	var blocked bool
	err := Stores.MySQLClient.QueryRow("SELECT blocked FROM users WHERE uid = ? LIMIT 1", userID).Scan(&blocked)
	if err != nil {
		return true
	}
	return blocked
}

func CloseAllSessions(userID uint32) {
	_, err := Stores.MySQLClient.Exec("DELETE FROM activities WHERE uid = ?", userID)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("CloseAllSessions failed: %s", err.Error()))
		CloseAllSessions(userID)
		return
	}
}

func CloseSession(userID uint32, refreshID uint16) {
	_, err := Stores.MySQLClient.Exec("DELETE FROM activities WHERE uid = ? AND refresh = ?", userID, refreshID)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("CloseSession failed: %s", err.Error()))
		CloseSession(userID, refreshID)
		return
	}
}

func SessionExists(userID uint32, refreshID uint16) bool {
	var blocked bool
	err := Stores.MySQLClient.QueryRow("SELECT blocked FROM activities WHERE uid = ? AND refresh = ?", userID, refreshID).Scan(&blocked)
	if err != nil {
		return true
	}
	return blocked
}

func UserHasPassword(userID uint32) bool {
	var hasPassword bool
	err := Stores.MySQLClient.QueryRow(`SELECT password IS NOT NULL AND password <> '' FROM users WHERE uid = ? LIMIT 1`, userID).Scan(&hasPassword)
	if err != nil {
		return false
	}
	return hasPassword
}

func PasswordIsStrong(password string) bool {
	n := len(password)
	if n < 7 || n > 20 {
		return false
	}
	var hasUpper, hasLower, hasDigit bool
	for i := 0; i < n; i++ {
		c := password[i]
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
		if hasUpper && hasLower && hasDigit {
			return true
		}
	}
	return hasUpper && hasLower && hasDigit
}

func UpdatePassword(userID uint32, password string) bool {
	if PasswordIsStrong(password) {
		_, err := Stores.MySQLClient.Exec(`UPDATE users SET password = ? WHERE uid = ? LIMIT 1`, _HashPassword(password), userID)
		if err != nil {
			Logger.AccidentalFailure(fmt.Sprintf("UpdatePassword failed: %s", err.Error()))
			return false
		}
		CloseAllSessions(userID)
		return true
	}
	return false
}

func PasswordMatches(userID uint32, password string) bool {
	if PasswordIsStrong(password) {
		var hash string
		err := Stores.MySQLClient.QueryRow(`SELECT password FROM users WHERE uid = ? LIMIT 1`, userID).Scan(&hash)
		if err == nil {
			if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil {
				return true
			}
		} else {
			Logger.AccidentalFailure(fmt.Sprintf("Fetch password hash failed: %s", err.Error()))
		}
	}
	return false
}

func _HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Hash Password failed: %s", err.Error()))
		return _HashPassword(password)
	}
	return string(hash)
}

func RecordNewUser(userID uint32, password string, mail string, name string) bool {
	var hash string
	if password != "" {
		hash = _HashPassword(password)
	}
	_, err := Stores.MySQLClient.Exec(
		"INSERT INTO users VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID,
		UserTypes.All.Viewer.Short,
		mail,
		name,
		false,
		hash,
		time.Now(),
	)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("RecordNewUser failed [%d-%s]: %s", userID, mail, err.Error()))
		return false
	}
	return true
}

func RecordReturningUser(refreshID uint16, userID uint32, rememberMe bool) bool {
	_, err := Stores.MySQLClient.Exec(
		"INSERT INTO activities VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		userID,
		refreshID,
		1,
		"",
		false,
		rememberMe,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("RecordReturningUser failed [%d-%d]: %s", userID, refreshID, err.Error()))
		return false
	}
	return true
}

func ServeAccountDetails() {
	channel := Stores.RedisClient.Subscribe(Stores.Ctx, Config.AccountDetailsRequestChannel).Channel()
	for message := range channel {
		var request TokenModels.AccountDetailsRequest
		if err := json.Unmarshal([]byte(message.Payload), &request); err != nil {
			Logger.AccidentalFailure(fmt.Sprintf("ServeAccountDetails Unmarshal failed %s: %s", message, err.Error()))
			continue
		}
		var response TokenModels.AccountDetailsResponse
		err := Stores.MySQLClient.QueryRow("SELECT uid, type, email, name, creation FROM users WHERE uid = ? LIMIT 1", request.UserID).
			Scan(&response.UserID, &response.Type, &response.Email, &response.Name, &response.Creation)
		if err != nil {
			Logger.AccidentalFailure(fmt.Sprintf("ServeAccountDetails Parse from DB failed: %s", err.Error()))
			continue
		}
		Stores.RedisClient.Publish(Stores.Ctx, fmt.Sprintf("%s:%s", Config.AccountDetailsResponseChannel, request.ServerID), response)
	}
}
