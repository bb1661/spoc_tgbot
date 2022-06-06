package main

import (
	"database/sql"
	"fmt"
	"time"
)

var db1 *sql.DB

func rcdb(conString, botUrl string) int {
	_ = sendMessage(botUrl, "Error ping db", 261609763)
	errdb := true
	for errdb {
		timerdb := time.NewTimer(time.Second * 10)
		<-timerdb.C
		db1, err := sql.Open("mssql", conString)
		if err == nil {
			errdb = false
			db1.Close()
		}

		fmt.Println("Error ping db")
		fmt.Scanf(" ")
	}
	_ = sendMessage(botUrl, "Reconnected to db", 261609763)

	fmt.Println("reconnected to db")
	return 1
}
