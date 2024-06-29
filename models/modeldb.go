package models

import (
	"database/sql"
	"fmt"
	"itsky/a2b-api-go/env"

	_ "github.com/go-sql-driver/mysql"
)

func getDB() *sql.DB {
	var connStr string
	if env.Env.DbPassword == "" {
		connStr = fmt.Sprintf("%s@tcp(%s:%s)/%s", env.Env.DbUser, env.Env.DbHost, env.Env.DbPort, env.Env.DbName)
	} else {
		connStr = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			env.Env.DbUser, env.Env.DbPassword, env.Env.DbHost, env.Env.DbPort, env.Env.DbName)
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return db
}

func GetCard(username string) *Card {
	db := getDB()
	defer db.Close()

	result := db.QueryRow("SELECT useralias,credit FROM cc_card WHERE useralias = ?", username)

	client := Card{}
	err := result.Scan(&client.Useralias, &client.Credit)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &client
}

// TODO(TobaniEG): create corresponding cc_logrefill & cc_logpayment records
func CardRecharge(username string, amount int) *Card {
	db := getDB()
	defer db.Close()

	_, err := db.Exec(
		"UPDATE cc_card SET credit = credit + ? WHERE useralias = ?",
		amount, username)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	result := db.QueryRow("SELECT useralias,credit FROM cc_card WHERE useralias = ?", username)
	client := Card{}
	err = result.Scan(&client.Useralias, &client.Credit)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &client
}
