package routes

import (
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	ExecutionController "data-fetcher-api/src/Controllers/Guest/Execution"
	HistoricalExecutionController "data-fetcher-api/src/Controllers/Guest/Execution/Historical"
	HistoricalExecutionTemplateController "data-fetcher-api/src/Controllers/Guest/Execution/Historical/Templates"
	ExecutionTemplateController "data-fetcher-api/src/Controllers/Guest/Execution/Templates"
	InstrumentController "data-fetcher-api/src/Controllers/Guest/Instrument"
	PerformanceController "data-fetcher-api/src/Controllers/Guest/Performance"
	HistoricalPerformanceController "data-fetcher-api/src/Controllers/Guest/Performance/Historical"
	HistoricalPerformanceTemplateController "data-fetcher-api/src/Controllers/Guest/Performance/Historical/Templates"
	PerformanceTemplateController "data-fetcher-api/src/Controllers/Guest/Performance/Templates"
	ProfitTemplatesController "data-fetcher-api/src/Controllers/Guest/Profit/Templates"
	ReliableScoreController "data-fetcher-api/src/Controllers/Guest/Reliable/Templates"
)

func SetupRouter() *gin.Engine {

	router := gin.Default()
	router.Use(gin.Logger())
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://127.0.0.1",
		"http://localhost",
		"http://127.0.0.1:3000",
		"http://localhost:3000",
		"http://127.0.0.1:8080",
		"http://localhost:8080",
	}
	config.ExposeHeaders = []string{"Link"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "OPTIONS"}
	config.AllowHeaders = []string{"Content-Type"}

	corsMiddleware := cors.New(config)

	router.Use(corsMiddleware, sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	router.GET("/guest/executions", ExecutionController.GetExecutions)
	router.GET("/guest/executions/historical", HistoricalExecutionController.GetExecutions)
	router.GET("/guest/instruments/:instrumentName/price", InstrumentController.GetInstrumentPrice)
	router.GET("/guest/performances", PerformanceController.GetPerformances)
	router.GET("/guest/performances/historical", HistoricalPerformanceController.GetPerformances)
	router.GET("/guest/templates/:id/executions/historical", HistoricalExecutionTemplateController.GetExecution)
	router.GET("/guest/templates/:id/executions", ExecutionTemplateController.GetExecution)
	router.GET("/guest/templates/eligible/estimation", ProfitTemplatesController.GetEstimatedProfitEligibleStrategies)
	router.GET("/guest/templates/:id/profit/estimation", ProfitTemplatesController.GetEstimatedProfit)
	router.GET("/guest/templates/:id/performances/historical", HistoricalPerformanceTemplateController.GetPerformance)
	router.GET("/guest/templates/:id/performances", PerformanceTemplateController.GetPerformance)
	router.GET("/guest/templates/:id/score", ReliableScoreController.GetReliableScore)

	return router
}
