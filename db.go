package main

import (
	"database/sql"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	var err error
	db, err = GetDatabase()
	if errorExists(err) {
		throwConnectionError(err)
	}
	return db
}

func GetDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", DB_HOST, USER, PASSWORD, DB_NAME, PORT)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return db, err
}

func CloseDB() {
	sqlDb, err := db.DB()
	if errorExists(err) {
		throwCloseError(err)
	} else {
		closeSqlDB(sqlDb)
	}
}

func throwConnectionError(err error) {
	fmt.Printf("Unexpected connection error: %v", err)
	os.Exit(3)
}

func throwCloseError(err error) {
	fmt.Printf("Unexpected error while closing: %v", err)
	os.Exit(3)
}

func closeSqlDB(sqlDb *sql.DB) {
	sqlDb.Close()
}
