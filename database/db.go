package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"log"
)

//todo: create wrappers for queryx, query, countx, etc that call GetDB

var database *sqlx.DB

func GetDB() *sqlx.DB {

	if database == nil {
		// connect to database
		connStr := "postgres://"+viper.GetString("db_username")+":"+viper.GetString("db_password")+"@"+viper.GetString("db_host")+"/"+viper.GetString("db_name")+"?sslmode=disable"
		fmt.Println(connStr)
		db, err := NewDB(connStr)
		if err != nil {
			log.Fatal(err)
		}
		database = db
	}

	return database
}

func NewDB(connectionString string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
