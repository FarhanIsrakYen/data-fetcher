package main

import (
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
	"log"
	"strconv"
	"data-fetcher-api/config/packages"
	"data-fetcher-api/src/Model"
	"data-fetcher-api/src/Repository"
	"time"
)

const DATA_ORDER = "ASC"
const EMPTY_VALUE = 0

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
	fileErr := packages.RemoveCacheFile()
	if fileErr != nil {
		sentry.CaptureException(fileErr)
	}
	performanceErr := StorePerformance()
	if performanceErr != nil {
		sentry.CaptureException(performanceErr)
	}
	historicalPerformanceErr := StoreHistoricalPerformance()
	if historicalPerformanceErr != nil {
		log.Println(historicalPerformanceErr)
		sentry.CaptureException(historicalPerformanceErr)
	}
	if err := packages.SaveCacheToFile(); err != nil {
		sentry.CaptureException(err)
	}
}

func StorePerformance() error {
	c := packages.CacheDeclare(true)
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)
	templateIds, err := Repository.GetPerformancesTemplateId(db)
	if err != nil {
		sentry.CaptureException(err)
	}

	for _, templateId := range templateIds {
		performances, avgProfit, avgMaxDrawDown, err :=
			Repository.GetPerformanceByTemplateId(templateId, DATA_ORDER, db)
		if err != nil {
			return err
		}
		data := make(map[string]interface{})
		data["avgMaxDrawDown"] = avgMaxDrawDown
		data["avgProfit"] = avgProfit
		data["score"] = getReliableScoreByTemplate(templateId)
		data["performances"] = performances
		var key = "performances" + "-" + strconv.Itoa(templateId)
		c.Set(key, data, cache.DefaultExpiration)
	}
	log.Println("Performance stored in cache successfully")
	return nil
}

func StoreHistoricalPerformance() error {
	c := packages.CacheDeclare(true)
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)
	templateIds, err := Repository.GetHistoricalPerformancesTemplateId(db)
	if err != nil {
		sentry.CaptureException(err)
	}
	for _, templateId := range templateIds {
		performances, avgProfit, avgMaxDrawDown, err :=
			Repository.GetHistoricalPerformanceByTemplateId(templateId, DATA_ORDER, db)
		if err != nil {
			return err
		}

		data := make(map[string]interface{})
		data["avgMaxDrawDown"] = avgMaxDrawDown
		data["avgProfit"] = avgProfit
		data["performances"] = performances
		var key = "performances" + "-" + "historical" + "-" + strconv.Itoa(templateId)
		c.Set(key, data, cache.DefaultExpiration)
	}
	log.Println("Historical Performance stored in cache successfully")
	return nil
}

func getReliableScoreByTemplate(templateId int) *int {
	var score int
	realTimeMetrics, _ := Repository.GetReliableDataByTemplateId(templateId, Repository.TYPE_REALTIME)
	backtestingMetrics, _ := Repository.GetReliableDataByTemplateId(templateId, Repository.TYPE_BACKTESTING)
	if len(realTimeMetrics) == EMPTY_VALUE || len(backtestingMetrics) == EMPTY_VALUE {
		return nil
	}
	reliableScore := Model.GetReliableMetricsScore(realTimeMetrics, backtestingMetrics)
	score = Model.GetReliableScore(reliableScore)
	return &score
}
