package main

import (
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/url"
	"os"
	"strings"
	"data-fetcher-api/config/packages"
	"data-fetcher-api/src/Repository"
	"time"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("env file not loaded")
	}
}

func main() {
	cloudDb := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(cloudDb)
	dockerDb := connectDockerTimescaleDb()
	defer closeDockerDatabaseConnection(dockerDb)
	packages.SentryInit()
	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()

	executions, historicalExecutions, performances, historicalPerformances, err := getData(dockerDb)
	if err != nil {
		log.Println(err)
		return
	}

	importData(cloudDb, executions, historicalExecutions, performances, historicalPerformances)
	log.Println("Data moved to cloud database successfully")
	return
}

func getData(dockerDb *gorm.DB) ([]*Repository.DfDataExecution,
	[]*Repository.DfDataHistoricalExecution,
	[]*Repository.DfDataPerformance,
	[]*Repository.DfDataHistoricalPerformance, error) {
	executions, executionsErr := getExecutions(dockerDb)
	if executionsErr != nil {
		return nil, nil, nil, nil, executionsErr
	}
	historicalExecutions, historicalExecutionsErr := getHistoricalExecutions(dockerDb)
	if historicalExecutionsErr != nil {
		return executions, nil, nil, nil, historicalExecutionsErr
	}
	performances, performancesErr := getPerformances(dockerDb)
	if performancesErr != nil {
		return executions, historicalExecutions, nil, nil, performancesErr
	}
	historicalPerformances, historicalPerformancesErr := getHistoricalPerformances(dockerDb)
	if historicalPerformancesErr != nil {
		return executions, historicalExecutions, performances, nil, historicalPerformancesErr
	}

	return executions,
		historicalExecutions,
		performances,
		historicalPerformances,
		nil
}
func importData(cloudDb *gorm.DB,
	executions []*Repository.DfDataExecution,
	historicalExecutions []*Repository.DfDataHistoricalExecution,
	performances []*Repository.DfDataPerformance,
	historicalPerformances []*Repository.DfDataHistoricalPerformance) {

	cloudExecutionData, _ := getExecutions(cloudDb)
	if len(cloudExecutionData) == 0 {
		Repository.CreateMultipleExecutions(executions)
	}
	cloudHistoricalExecutionData, _ := getHistoricalExecutions(cloudDb)
	if len(cloudHistoricalExecutionData) == 0 {
		Repository.CreateMultipleHistoricalExecutions(historicalExecutions)
	}
	cloudPerformanceData, _ := getPerformances(cloudDb)
	if len(cloudPerformanceData) == 0 {
		insertPerformances(cloudDb, performances)
	}
	cloudHistoricalPerformanceData, _ := getHistoricalPerformances(cloudDb)
	if len(cloudHistoricalPerformanceData) == 0 {
		insertHistoricalPerformances(cloudDb, historicalPerformances)
	}
}

func connectDockerTimescaleDb() *gorm.DB {

	dbURL := os.Getenv("TIMESCALEDB_PREV_CONNECTION_STRING")
	password := os.Getenv("TIMESCALEDB_PREV_PASSWORD")
	encodedPassword := url.QueryEscape(password)
	dbURL = strings.Replace(dbURL, password, encodedPassword, 1)

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})

	if err != nil {
		log.Println(err.Error())
	}

	if err != nil {
		log.Println(err.Error())
	}

	return db
}

func closeDockerDatabaseConnection(db *gorm.DB) {
	dbSQL, _ := db.DB()
	defer dbSQL.Close()
	return
}

func getExecutions(db *gorm.DB) ([]*Repository.DfDataExecution, error) {
	var executions []*Repository.DfDataExecution
	err := db.Find(&executions).Error
	if err != nil {
		return nil, err
	}
	return executions, nil
}

func getHistoricalExecutions(db *gorm.DB) ([]*Repository.DfDataHistoricalExecution, error) {
	var executions []*Repository.DfDataHistoricalExecution
	err := db.Find(&executions).Error
	if err != nil {
		return nil, err
	}
	return executions, nil
}

func getPerformances(db *gorm.DB) ([]*Repository.DfDataPerformance, error) {
	var performances []*Repository.DfDataPerformance
	err := db.Find(&performances).Error
	if err != nil {
		return nil, err
	}
	return performances, nil
}

func insertPerformances(db *gorm.DB, performances []*Repository.DfDataPerformance) error {
	err := db.Create(&performances).Error
	return err
}

func getHistoricalPerformances(db *gorm.DB) ([]*Repository.DfDataHistoricalPerformance, error) {
	var performances []*Repository.DfDataHistoricalPerformance
	err := db.Find(&performances).Error
	if err != nil {
		return nil, err
	}
	return performances, nil
}

func insertHistoricalPerformances(db *gorm.DB, performances []*Repository.DfDataHistoricalPerformance) error {
	err := db.Create(&performances).Error
	return err
}
