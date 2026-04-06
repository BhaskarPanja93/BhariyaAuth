package responses

import "time"

// All response types

type AccountDetailsRequestT struct {
	ServerID string `json:"sid"`
	UserID   int32  `json:"uid"`
}

type AccountDetailsResponseT struct {
	UserID  int32     `json:"uid"`
	Email   string    `json:"mail"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
}

type APIResponseT struct {
	Success       bool     `json:"success,omitempty"`
	Reply         any      `json:"reply,omitempty"`
	ModifyAuth    bool     `json:"modify-auth,omitempty"`
	NewToken      string   `json:"new-token,omitempty"`
	Notifications []string `json:"notifications,omitempty"`
	RetryAfter    int      `json:"retry-after,omitempty"`
}

type SingleUserActivityT struct {
	ID         string    `json:"id"`
	Count      int16     `json:"count"`
	Remembered bool      `json:"remembered"`
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
	OS         string    `json:"os"`
	Device     string    `json:"device"`
	Browser    string    `json:"browser"`
}

type UserActivityResponseT struct {
	User       string                `json:"user"`
	Refresh    string                `json:"refresh"`
	Activities []SingleUserActivityT `json:"activities"`
}
