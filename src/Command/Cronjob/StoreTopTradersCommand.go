package main

import (
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"log"
	"data-fetcher-api/config/packages"
	MqPublishLib "data-fetcher-api/src/Lib/MqPublish"
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
	err := StoreTopTraders()
	if err != nil {
		sentry.CaptureException(err)
	}
}

func StoreTopTraders() error {
	strategies, err := Repository.GetTopStrategiesId()

	if err != nil {
		sentry.CaptureException(err)
		fmt.Println(err)
	}

	if len(strategies) != 0 {
		topic := "/api/tc/data/templates/top-performance/create"
		jsonData, err := json.Marshal(strategies)
		if err != nil {
			fmt.Println("Failed to create user strategy profit")
			return nil
		}
		MqPublishLib.MqPublish(string(jsonData), topic)
	}
	log.Println("Top traders templates stored successfully")
	return nil
}
