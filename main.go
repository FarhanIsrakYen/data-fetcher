package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"data-fetcher-api/config/packages"
	"data-fetcher-api/config/routes"
	"data-fetcher-api/src/Helper"
	MqSubscribeLib "data-fetcher-api/src/Lib/MqSubscribe"
)

func main() {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("env file not loaded")
	}

	packages.SentryInit()

	MqSubscribeLib.MqSubscribe()

	router := routes.SetupRouter()

	parameter, err := Helper.GetParameter()
	if err != nil {
		panic(err)
	}
	gin.SetMode(parameter.Parameters.GinMode)
	router.Run(":" + parameter.Parameters.GinServerPort)
}
