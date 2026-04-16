package account

import (
	Config "BhariyaAuth/constants/config"
	Logs "BhariyaAuth/processors/logs"
	Stores "BhariyaAuth/stores"
	"time"
)

// DatabaseAutoVacuum periodically cleans up expired device/session records
// from the database to prevent stale session accumulation.
//
// This background worker continuously deletes device records that are older
// than the configured refresh token expiry duration. It ensures that:
//  1. Expired sessions (no longer refreshable) are removed.
//  2. Database size remains controlled over time.
//  3. Old device records do not impact query performance.
//
// Flow Summary:
//
//	compute expiry cutoff → delete stale records → sleep → repeat
//
// Cleanup Logic:
//   - A device is considered expired if:
//     updated < (current time - RefreshTokenExpireDelta)
//   - This aligns with refresh token lifecycle expiration.
//
// Retry Strategy:
// - On failure: retry after 1 minute (fast retry).
// - On success: run once every 24 hours (daily cleanup).
//
// Concurrency Model:
// - Runs as a long-lived background goroutine.
// - Assumes single instance execution (no distributed locking).
//
// Dependencies:
// - devices table with `updated` timestamp column.
// - Config.RefreshTokenExpireDelta defines session lifetime.
//
// Notes:
// - This does NOT immediately revoke active sessions.
// - Only removes sessions already expired by policy.
//
// Returns:
// - No return value (runs indefinitely).
func DatabaseAutoVacuum() {
	for {
		// Calculate cutoff time for expired sessions
		cutoff := time.Now().Add(-Config.RefreshTokenExpireDelta)

		// Delete all device records older than cutoff
		_, err := Stores.SQLClient.Exec(
			Config.CtxBG,
			"DELETE FROM devices WHERE updated < $1",
			cutoff,
		)

		if err != nil {
			// On failure: retry sooner to recover quickly
			Logs.RootLogger.Add(Logs.Error, "processors/account/autovacuum", "", "Cleanup failed: "+err.Error())
			time.Sleep(time.Minute)
			continue
		}

		// On success: run cleanup once per day
		time.Sleep(24 * time.Hour)
	}
}
