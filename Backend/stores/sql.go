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

// SQLClient is a global PostgreSQL connection pool.
//
// Notes:
// - Uses pgxpool for efficient connection pooling.
// - Safe for concurrent use across goroutines.
// - Initialized once via ConnectSQL().
var SQLClient *pgxpool.Pool

// ConnectSQL initializes a PostgreSQL connection pool with fallback and retry.
//
// Overview:
// This function attempts to establish a database connection using:
//  1. UNIX socket (preferred for local deployments).
//  2. TCP/IP fallback (for remote environments).
//
// Flow:
//
//	build DSN → parse config → create pool → ping → retry on failure
//
// Behavior:
// - Retries indefinitely until successful connection.
// - Alternates between socket and TCP on failure.
// - Uses connection pooling with tuned parameters.
//
// Configuration:
// - Credentials and connection details from Secrets.
// - Pool tuning via pgxpool.Config.
//
// Important:
// - Blocks indefinitely until DB becomes available.
// - Intended for controlled startup phase.
func ConnectSQL() {
	Logs.RootLogger.Add(Logs.Intent, sqlFileName, "", "Attempting connect SQL")

	// Prevent re-initialization
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

			panic(err) // configuration error → unrecoverable
		}

		config.MaxConns = 25                        // max concurrent DB connections
		config.MinConns = 5                         // minimum idle connections
		config.MaxConnLifetime = 30 * time.Minute   // recycle connections
		config.MaxConnIdleTime = 10 * time.Minute   // close idle connections
		config.HealthCheckPeriod = 30 * time.Second // background health checks

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

		// Success
		break
	}
}
