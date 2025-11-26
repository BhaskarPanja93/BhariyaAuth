package responses

import "time"

type AccountDetailsRequestT struct {
	ServerID string `json:"sid"`
	UserID   uint32 `json:"uid"`
}

type AccountDetailsResponseT struct {
	UserID   uint32    `json:"uid"`
	Type     string    `json:"typ"`
	Email    string    `json:"mail"`
	Name     string    `json:"name"`
	Creation time.Time `json:"creation"`
}

type APIResponseT struct {
	Success       bool     `json:"success"`
	Reply         any      `json:"reply"`
	ModifyAuth    bool     `json:"modify-auth"`
	NewToken      string   `json:"new-token"`
	Notifications []string `json:"notifications"`
	RetryAfter    int      `json:"retry-after"`
}

type SingleUserActivityT struct {
	ID         string    `json:"id"`
	Count      uint16    `json:"count"`
	Remembered bool      `json:"remembered"`
	Creation   time.Time `json:"creation"`
	Updated    time.Time `json:"updated"`
	OS         string    `json:"os"`
	Device     string    `json:"device"`
	Browser    string    `json:"browser"`
}

type UserActivityResponseT struct {
	UserID     string                `json:"user_id"`
	DeviceID   string                `json:"device_id"`
	Activities []SingleUserActivityT `json:"activities"`
}
