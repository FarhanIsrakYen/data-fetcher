package main

import (
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"log"
	"data-fetcher-api/config/packages"
	"data-fetcher-api/src/Model"
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
	packages.SentryInit()
	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()
	err := CreateHistoricalPerformance()
	if err != nil {
		sentry.CaptureException(err)
	}
}

func CreateHistoricalPerformance() error {

	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)

	performanceLog, _ := Repository.GetOldPerformanceLogData(db)
	if performanceLog.TemplateId != 0 {
		executionTime := performanceLog.Time
		year, month, day := executionTime.Date()

		startOfDay := time.Date(year, month, day, 0, 0, 0, 0, executionTime.Location())

		endOfDay := startOfDay.AddDate(0, 0, 1).Add(-time.Nanosecond)

		templateIds, err := Repository.GetHistoricalExecutionsTemplateId(db, startOfDay, endOfDay)
		if err != nil {
			return err
		}
		if len(templateIds) != 0 {
			for i := 0; i < len(templateIds); i++ {
				Model.CalculateHistoricalPerformance(db, templateIds[i], startOfDay, endOfDay)
			}
			Repository.UpdatePerformanceLog(db, executionTime)
			log.Println("Historical Performance data Generated")
		}
	}
	return nil
}
