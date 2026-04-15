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
	err := CreatePerformance()
	if err != nil {
		sentry.CaptureException(err)
	}
}

func CreatePerformance() error {

	now := time.Now().UTC()
	previousDay := now.AddDate(0, 0, -1)
	startTime := time.Date(previousDay.Year(), previousDay.Month(), previousDay.Day(), 0, 0, 0, 0, time.UTC)
	endTime := time.Date(previousDay.Year(), previousDay.Month(), previousDay.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)

	templateIds, err := Repository.GetExecutionsTemplateId(db, startTime, endTime)
	if err != nil {
		return err
	}
	if len(templateIds) != 0 {
		for i := 0; i < len(templateIds); i++ {
			Model.CalculatePerformance(db, templateIds[i], startTime, endTime)
		}
		Repository.UpdatePerformanceLog(db, startTime)
		log.Println("Performance Generated")
	}
	return nil
}
