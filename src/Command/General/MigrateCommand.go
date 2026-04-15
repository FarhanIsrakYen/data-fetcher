package main

import (
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"log"
	"data-fetcher-api/config/packages"
	"data-fetcher-api/src/Repository"
	"data-fetcher-api/src/Migrations"
	"time"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("env file not loaded")
	}
}

func main() {
	packages.SentryInit()
	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()

	db := packages.ConnectTimescaleDb()

	err := db.AutoMigrate(&Repository.DfDataExecution{},
		&Repository.DfDataPerformance{},
		&Repository.DfDataData{},
		&Repository.DfDataPerformanceLog{},
		&Repository.DfDataHistoricalExecution{},
		&Repository.DfDataHistoricalPerformance{},
		&Repository.DfDataInstrumentLog{},
		&Repository.DfDataReliable{},
	)

	// hyper table
	Migrations.CreateDataHyperTable(db)
	Migrations.CreateExecutionHyperTable(db)
	Migrations.CreatePerformanceHyperTable(db)
	Migrations.CreatePerformanceLogHyperTable(db)
	Migrations.CreateHistoricalExecutionHyperTable(db)
	Migrations.CreateHistoricalPerformanceHyperTable(db)
	Migrations.CreateReliableHyperTable(db)

	migrator := db.Migrator()
	if migrator.HasColumn(&Repository.DfDataPerformance{}, "is_historical") {
		migrator.DropColumn(&Repository.DfDataPerformance{}, "is_historical")
	}
	if migrator.HasColumn(&Repository.DfDataExecution{}, "is_historical") {
		migrator.DropColumn(&Repository.DfDataExecution{}, "is_historical")
	}
	if migrator.HasColumn(&Repository.DfDataHistoricalExecution{}, "position") {
		migrator.DropColumn(&Repository.DfDataHistoricalExecution{}, "position")
	}

	if err != nil {
		panic("Failed to Migrate")
		return
	}
	packages.CloseDatabaseConnection(db)

	println("Migrate Successfully")
	return
}
