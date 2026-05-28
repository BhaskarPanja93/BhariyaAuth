package account

import (
	Config "BhariyaAuth/constants/config"
	Logs "BhariyaAuth/processors/logs"
	Stores "BhariyaAuth/stores"
	"time"
)

func DatabaseAutoVacuum() {
	for {
		cutoff := time.Now().Add(-Config.RefreshTokenExpireDelta)

		_, err := Stores.SQLClient.Exec(
			Config.CtxBG,
			"DELETE FROM devices WHERE updated < $1",
			cutoff,
		)

		if err != nil {
			Logs.RootLogger.Add(Logs.Error, "processors/account/autovacuum", "", "Cleanup failed: "+err.Error())
			time.Sleep(time.Minute)
			continue
		}

		time.Sleep(24 * time.Hour)
	}
}
