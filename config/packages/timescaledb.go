package packages

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

func ConnectTimescaleDb() *gorm.DB {

	dbURL := os.Getenv("TIMESCALEDB_CONNECTION_STRING")
	password := os.Getenv("TIMESCALEDB_PASSWORD")
	encodedPassword := url.QueryEscape(password)
	dbURL = strings.Replace(dbURL, password, encodedPassword, 1)

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})

	if err != nil {
		log.Println(err.Error())
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Println(err.Error())
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db
}

func CloseDatabaseConnection(db *gorm.DB) {
	dbSQL, _ := db.DB()
	defer dbSQL.Close()
	return
}
