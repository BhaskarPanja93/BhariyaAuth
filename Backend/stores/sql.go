package stores

import (
	Config "BhariyaAuth/constants/config"
	"fmt"
	"time"

	Secrets "BhariyaAuth/constants/secrets"

	"github.com/jackc/pgx/v5/pgxpool"
)

var SQLClient *pgxpool.Pool

func ConnectSQL() {
	if SQLClient != nil {
		return
	}

	var useSocket = true

	for {
		var dsn string
		if useSocket {
			fmt.Println("Trying SQL via UNIX socket...")
			dsn = fmt.Sprintf(
				"postgres://%s:%s@/%s?host=%s&sslmode=disable&TimeZone=UTC",
				Secrets.SQLUser,
				Secrets.SQLPassword,
				Secrets.SQLDBName,
				Secrets.SQLSocket,
			)
		} else {
			fmt.Println("Trying SQL via TCP/IP...")
			dsn = fmt.Sprintf(
				"postgres://%s:%s@%s:%s/%s?sslmode=disable&TimeZone=UTC",
				Secrets.SQLUser,
				Secrets.SQLPassword,
				Secrets.SQLHost,
				Secrets.SQLPort,
				Secrets.SQLDBName,
			)
		}

		config, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			panic(err)
		}
		config.MaxConns = 10
		config.MinConns = 1
		config.MaxConnLifetime = 15 * time.Minute
		config.MaxConnIdleTime = 5 * time.Minute
		config.HealthCheckPeriod = 1 * time.Minute

		SQLClient, err = pgxpool.NewWithConfig(Config.CtxBG, config)
		if err != nil {
			fmt.Println("Failed to create SQL connection:", err.Error())
			useSocket = !useSocket
			time.Sleep(2 * time.Second)
			continue
		}

		err = SQLClient.Ping(Config.CtxBG)
		if err != nil {
			fmt.Println("Cannot ping SQL:", err.Error())
			SQLClient.Close()
			SQLClient = nil
			useSocket = !useSocket
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	fmt.Println("SQL connection established successfully")
}
