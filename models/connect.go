package models

import (
	"database/sql"
	"fmt"
	"itsky/a2b-api-go/env"
	"itsky/a2b-api-go/utils"
)

// only use for testing
func SetDB(s *sql.DB) { db = s }

func ConnectDB() error {
	var connStr string
	if env.Env.DbPassword == "" {
		connStr = fmt.Sprintf("%s@tcp(%s:%s)/%s", env.Env.DbUser, env.Env.DbHost, env.Env.DbPort, env.Env.DbName)
	} else {
		connStr = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			env.Env.DbUser, env.Env.DbPassword, env.Env.DbHost, env.Env.DbPort, env.Env.DbName)
	}
	var err error
	db, err = sql.Open("mysql", connStr)
	return err
}

func DisconnectDB() {
	if err := db.Close(); err != nil {
		utils.Log.Println(err)
	}
}
