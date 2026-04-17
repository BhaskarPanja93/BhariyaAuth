package account

import (
	Config "BhariyaAuth/constants/config"
	Stores "BhariyaAuth/stores"
	"errors"
	"strconv"
)

func denySingleDeviceAccess(userID int32, deviceID int16) error {
	key := Config.RefreshBlocked + ":" + strconv.Itoa(int(userID)) + strconv.Itoa(int(deviceID))
	err := Stores.RedisClient.Set(Config.CtxBG, key, nil, Config.AccessTokenExpireDelta).Err()
	if err != nil {
		return errors.New("Deny single device access - redis set: " + err.Error())
	}
	return err
}

func CheckDeviceAccessDenied(userID int32, deviceID int16) (bool, error) {
	key := Config.RefreshBlocked + ":" + strconv.Itoa(int(userID)) + strconv.Itoa(int(deviceID))
	exists, err := Stores.RedisClient.Exists(Config.CtxBG, key).Result()
	if err != nil {
		return true, errors.New("Check single device access - redis exists: " + err.Error())
	}
	return exists == 1, err
}

func DenyAllDevicesFromRenewing(userID int32) error {
	rows, err := Stores.SQLClient.Query(
		Config.CtxBG,
		"DELETE FROM devices WHERE user_id = $1 RETURNING device_id",
		userID)
	if err != nil {
		return errors.New("Deny all device renew - SQL query: " + err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var deviceID int16
		err = rows.Scan(&deviceID)
		if err != nil {
			return errors.New("Deny all device renew - SQL scan loop: " + err.Error())
		}
		err = denySingleDeviceAccess(userID, deviceID)
		if err != nil {
			return errors.New("Deny all device renew: " + err.Error())
		}
	}
	err = rows.Err()
	if err != nil {
		return errors.New("Deny all device renew - SQL err: " + err.Error())
	}
	return nil
}

func DenySingleDeviceFromRenewing(userID int32, deviceID int16) error {
	_, err := Stores.SQLClient.Exec(
		Config.CtxBG,
		"DELETE FROM devices WHERE user_id = $1 AND device_id = $2 ",
		userID, deviceID)
	if err != nil {
		return errors.New("Deny single device renew - SQL exec: " + err.Error())
	}
	err = denySingleDeviceAccess(userID, deviceID)
	if err != nil {
		return errors.New("Deny single device renew: " + err.Error())
	}
	return nil
}

func BlacklistUser(userID int32) error {
	err := DenyAllDevicesFromRenewing(userID)
	if err != nil {
		return errors.New("Blacklist user: " + err.Error())
	}
	_, err = Stores.SQLClient.Exec(
		Config.CtxBG,
		"UPDATE users SET blocked = TRUE WHERE user_id = $1", userID)
	if err != nil {
		return errors.New("Blacklist user - SQL exec: " + err.Error())
	}
	return nil
}
