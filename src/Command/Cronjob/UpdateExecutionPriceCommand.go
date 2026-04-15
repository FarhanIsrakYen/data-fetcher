package main

import (
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"log"
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
	packages.SentryInit()
	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()
	executionErr := UpdateExecutionPrice()
	if executionErr != nil {
		sentry.CaptureException(executionErr)
	}

	historicalExecutionErr := UpdateHistoricalExecutionPrice()
	if historicalExecutionErr != nil {
		sentry.CaptureException(historicalExecutionErr)
	}
}

func UpdateExecutionPrice() error {

	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)
	executions, dataErr := Repository.GetExecutionWithoutPrice(db)
	if dataErr != nil {
		return dataErr
	}
	if len(executions) != 0 {
		for i := 0; i < len(executions); i++ {
			data, err := Repository.GetNearestData(db, executions[i].Instrument, executions[i].Time)

			if err == nil {
				_, err := Repository.UpdateExecution(
					db,
					executions[i].TemplateId,
					executions[i].Signals,
					executions[i].Position,
					executions[i].Instrument,
					&data.Open,
					executions[i].Time)
				if err != nil {
					return err
				}
			}
		}
		log.Println("Execution price updated")
	}
	return nil
}

func UpdateHistoricalExecutionPrice() error {
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)
	executions, dataErr := Repository.GetHistoricalExecutionWithoutPrice(db)
	if dataErr != nil {
		return dataErr
	}
	if len(executions) != 0 {
		for i := 0; i < len(executions); i++ {
			data, err := Repository.GetNearestData(db, executions[i].Instrument, executions[i].Time)

			if err == nil {
				_, err := Repository.UpdateHistoricalExecution(
					db,
					executions[i].TemplateId,
					executions[i].Signals,
					executions[i].Instrument,
					&data.Open,
					executions[i].Time)
				if err != nil {
					return err
				}
			}
		}
		log.Println("Historical Execution price updated")
	}
	return nil
}
