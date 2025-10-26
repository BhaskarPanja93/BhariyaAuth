package stores

import (
	Secrets "BhariyaAuth/constants/secrets"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var MySQLClient *sql.DB

func ConnectMySQL() {
	if MySQLClient != nil {
		return
	}

	var (
		err       error
		useSocket = true
	)

	for {
		var dsn string
		if useSocket {
			fmt.Println("Trying MySQL via UNIX socket...")
			dsn = fmt.Sprintf("%s:%s@unix(%s)/%s?parseTime=true",
				Secrets.MySQLUser,
				Secrets.MySQLPassword,
				Secrets.MySQLSocket,
				Secrets.MySQLDBName,
			)
		} else {
			fmt.Println("Trying MySQL via TCP/IP...")
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
				Secrets.MySQLUser,
				Secrets.MySQLPassword,
				Secrets.MySQLHost,
				Secrets.MySQLPort,
				Secrets.MySQLDBName,
			)
		}

		MySQLClient, err = sql.Open("mysql", dsn)
		if err != nil {
			fmt.Println("Failed to create MySQL connection:", err.Error())
			useSocket = !useSocket
			time.Sleep(2 * time.Second)
			continue
		}

		err = MySQLClient.Ping()
		if err != nil {
			fmt.Println("Cannot ping MySQL:", err.Error())
			err = MySQLClient.Close()
			if err != nil {
			}
			MySQLClient = nil
			useSocket = !useSocket
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	MySQLClient.SetMaxOpenConns(10)
	MySQLClient.SetMaxIdleConns(2000)
	MySQLClient.SetConnMaxLifetime(5 * time.Minute)

	fmt.Println("MySQL connection established successfully")
}
