package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"data-fetcher-api/src/Helper"
	"data-fetcher-api/config/packages"
	"github.com/getsentry/sentry-go"
	"time"
)

const (
	envPath    = ".env"
	configPath = "./config/google/config.json"
)

var keysToExtractFromEnv = []string{
	"GOOGLE_CLIENT_ID",
	"GOOGLE_PRIVATE_KEY",
	"GOOGLE_PRIVATE_KEY_ID",
}

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
	convertEnvToConfig()
	copyYamlToJson()
}

func convertEnvToConfig() {
	envData := make(map[string]string)
	for _, key := range keysToExtractFromEnv {
		removableKey := strings.Split(strings.ToLower(key), "_")
		lowercaseKey := strings.Replace(strings.ToLower(key), removableKey[0]+"_", "", 1)
		envData[lowercaseKey] = os.Getenv(key)
	}

	jsonData, err := json.MarshalIndent(envData, "", "  ")
	if err != nil {
		panic(err)
	}
	
	err = ioutil.WriteFile(configPath, jsonData, 0644)
	if err != nil {
		panic(err)
	}
}

func copyYamlToJson() {
	jsonFile, err := os.OpenFile(configPath, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			panic(err)
		}
	}(jsonFile)

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	var jsonMap map[string]string
	err = json.Unmarshal(jsonData, &jsonMap)
	if err != nil {
		panic(err)
	}

	parameter, err := Helper.GetParameter()
	if err != nil {
		panic(err)
	}
	jsonMap["token_uri"] = parameter.Parameters.TokenUri
	jsonMap["type"] = parameter.Parameters.Type
	jsonMap["auth_provider_x509_cert_url"] = parameter.Parameters.AuthProviderCertUrl
	jsonMap["auth_uri"] = parameter.Parameters.AuthUri
	jsonMap["universe_domain"] = parameter.Parameters.UniverseDomain
	jsonMap["project_id"] = parameter.Parameters.ProjectId
	jsonMap["client_email"] = parameter.Parameters.ClientEmail
	jsonMap["client_x509_cert_url"] = parameter.Parameters.ClientCertUrl

	mergedJsonData, err := json.MarshalIndent(jsonMap, "", "  ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(configPath, mergedJsonData, 0644)
	if err != nil {
		panic(err)
	}
	log.Println("JSON file for google service created successfully!")
	os.Exit(0)
}
