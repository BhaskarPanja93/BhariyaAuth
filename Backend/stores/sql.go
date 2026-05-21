package stores

import (
	Config "BhariyaAuth/constants/config"
	Logs "BhariyaAuth/processors/logs"
	"fmt"
	"time"

	Secrets "BhariyaAuth/constants/secrets"

	"github.com/jackc/pgx/v5/pgxpool"
)

const sqlFileName = "stores/sql"

var SQLClient *pgxpool.Pool

func ConnectSQL() {
	Logs.RootLogger.Add(Logs.Intent, sqlFileName, "", "Attempting connect SQL")

	if SQLClient != nil {
		return
	}

	var useSocket = true

	for {

		var dsn string

		if useSocket {
			Logs.RootLogger.Add(Logs.Intent, sqlFileName, "", "SQL using unix socket")

			dsn = fmt.Sprintf("postgres://%s:%s@/%s?host=%s&port=%s&sslmode=disable&TimeZone=UTC", Secrets.SQLUser, Secrets.SQLPassword, Secrets.SQLDBName, Secrets.SQLSocket, Secrets.SQLPort)

		} else {
			Logs.RootLogger.Add(Logs.Intent, sqlFileName, "", "SQL using TCP")

			dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&TimeZone=UTC", Secrets.SQLUser, Secrets.SQLPassword, Secrets.SQLHost, Secrets.SQLPort, Secrets.SQLDBName)
		}

		config, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			Logs.RootLogger.Add(Logs.Error, sqlFileName, "", "SQL Config parse error: "+err.Error())

			panic(err)
		}

		config.MaxConns = 25
		config.MinConns = 5
		config.MaxConnLifetime = 30 * time.Minute
		config.MaxConnIdleTime = 10 * time.Minute
		config.HealthCheckPeriod = 30 * time.Second

		SQLClient, err = pgxpool.NewWithConfig(Config.CtxBG, config)
		if err != nil {
			Logs.RootLogger.Add(Logs.Error, sqlFileName, "", "SQL Connect failed: "+err.Error())

			useSocket = !useSocket
			time.Sleep(2 * time.Second)
			continue
		}
		Logs.RootLogger.Add(Logs.Info, sqlFileName, "", "SQL Connected")

		err = SQLClient.Ping(Config.CtxBG)
		if err != nil {
			Logs.RootLogger.Add(Logs.Error, sqlFileName, "", "SQL Ping failed: "+err.Error())

			SQLClient.Close()
			SQLClient = nil

			useSocket = !useSocket
			time.Sleep(2 * time.Second)
			continue
		}
		Logs.RootLogger.Add(Logs.Info, sqlFileName, "", "SQL Connected and Pinged")

		break
	}
}
