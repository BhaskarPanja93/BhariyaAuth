package account

import "errors"

// Custom errors

var (
	DataRequestTimedOutError = errors.New("account data request: timed out")
)
