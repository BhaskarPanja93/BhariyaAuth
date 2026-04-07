package account

import (
	Config "BhariyaAuth/constants/config"
	Stores "BhariyaAuth/stores"
	"fmt"
)

func denySingeDeviceAccess(userID int32, deviceID int16) error {
	key := fmt.Sprintf("%s:%d:%d", Config.RefreshBlocked, userID, deviceID)
	return Stores.RedisClient.Set(Config.CtxBG, key, nil, Config.AccessTokenExpireDelta).Err()
}

func CheckDeviceAccessDenied(userID int32, deviceID int16) (bool, error) {
	key := fmt.Sprintf("%s:%d:%d", Config.RefreshBlocked, userID, deviceID)
	exists, err := Stores.RedisClient.Exists(Config.CtxBG, key).Result()
	if err != nil {
		return true, err
	}
	return exists == 1, nil
}

func DenyAllDevicesFromRenewing(userID int32) error {
	rows, err := Stores.SQLClient.Query(
		Config.CtxBG,
		"SELECT device_id FROM devices WHERE user_id = $1",
		userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var deviceID int16
		err = rows.Scan(&deviceID)
		if err != nil {
			return err
		}
		err = denySingeDeviceAccess(userID, deviceID)
		if err != nil {
			return err
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	_, err = Stores.SQLClient.Exec(
		Config.CtxBG,
		"DELETE FROM devices WHERE user_id = $1",
		userID)
	return err
}

func DenySingleDeviceFromRenewing(userID int32, deviceID int16) error {
	_, err := Stores.SQLClient.Exec(
		Config.CtxBG,
		"DELETE FROM devices WHERE user_id = $1 AND device_id = $2",
		userID, deviceID)
	if err != nil {
		return err
	}
	return denySingeDeviceAccess(userID, deviceID)
}

func BlacklistUser(userID int32) error {
	err := DenyAllDevicesFromRenewing(userID)
	if err != nil {
		return err
	}
	_, err = Stores.SQLClient.Exec(
		Config.CtxBG,
		"UPDATE users SET blocked = TRUE WHERE user_id = $1", userID)
	return err
}
