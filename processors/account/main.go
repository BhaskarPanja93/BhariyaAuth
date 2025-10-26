package account

import (
	TokenModels "BhariyaAuth/models/tokens"
	UserTypes "BhariyaAuth/models/users"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"time"

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
	_, err := Stores.MySQLClient.Exec("UPDATE users SET blocked = ? WHERE uid = ?", true, userID)
	if err != nil {
		return
	}
	rows, err := Stores.MySQLClient.Query("SELECT refresh FROM activities WHERE uid = ? AND blocked = ?", userID, false)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var refreshID uint16
		if err = rows.Scan(&refreshID); err != nil {
			return
		}
		TokenProcessor.BlacklistRefresh(userID, refreshID, false)
	}
	_, err = Stores.MySQLClient.Exec("UPDATE activities SET blocked = ? WHERE uid = ? AND blocked = ?", true, userID, false)
	if err != nil {
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
	if n < 7 || n > 49 {
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
		hash := _HashPassword(password)
		_, err := Stores.MySQLClient.Exec(`UPDATE users SET password = ? WHERE uid = ? LIMIT 1`, hash, userID)
		if err != nil {
			return false
		}
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
		}
	}
	return false
}

func _HashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func RecordNewUser(userID uint32, SignUpData TokenModels.SignUpT) bool {
	var hash string
	if SignUpData.Password != "" {
		hash = _HashPassword(SignUpData.Password)
	}
	_, err := Stores.MySQLClient.Exec(
		"INSERT INTO users VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		userID,
		UserTypes.All.Viewer.Short,
		SignUpData.Mail,
		SignUpData.First,
		SignUpData.Last,
		false,
		hash,
		time.Now(),
	)
	if err != nil {
		return false
	}
	return true
}

func RecordReturningUser(refreshID uint16, SignInData TokenModels.SignInT) bool {
	_, err := Stores.MySQLClient.Exec(
		"INSERT INTO activities VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		SignInData.UserID,
		refreshID,
		1,
		"",
		false,
		SignInData.RememberMe,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return false
	}
	return true
}
