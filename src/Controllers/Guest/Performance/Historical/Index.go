package HistoricalPerformanceController

import (
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"data-fetcher-api/src/Repository"
	"time"
)

type HistoricalPerformance struct {
	TemplateId       int       `json:"templateId"`
	Profit           float32   `json:"profit" `
	MaxDrawdown      float32   `json:"maxDrawdown" `
	Trades           int       `json:"trades"`
	WinningTrades    int       `json:"winningTrades"`
	LosingTrades     int       `json:"losingTrades"`
	CumulativeProfit float32   `json:"cumulativeProfit"`
	Time             time.Time `json:"time"`
}

func GetPerformances(ctx *gin.Context) {
	limitStr := ctx.Query("limit")
	limit, _ := strconv.Atoi(limitStr)
	orderBy := ctx.Query("orderBy")
	order := ctx.Query("order")
	pageStr := ctx.Query("page")
	page, _ := strconv.Atoi(pageStr)
	filters := ctx.Request.URL.Query()

	performances, pagination, err := Repository.GetHistoricalPerformances(limit, orderBy, order, page, filters)

	if err != nil {
		sentry.CaptureException(err)
	}

	if len(performances) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  true,
			"message": "No historical performance found",
			"data":    []interface{}{},
		})
		return
	}
	var data []HistoricalPerformance
	var cumulativeProfitPercentage float32 = 0
	for i, _ := range performances {
		cumulativeProfitPercentage += performances[i].Profit
		data = append(data, HistoricalPerformance{
			TemplateId:       performances[i].TemplateId,
			Profit:           performances[i].Profit,
			MaxDrawdown:      performances[i].MaxDrawdown,
			Trades:           performances[i].Trades,
			WinningTrades:    performances[i].WinningTrades,
			LosingTrades:     performances[i].LosingTrades,
			CumulativeProfit: cumulativeProfitPercentage,
			Time:             performances[i].Time,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":     true,
		"message":    "List of historical performances",
		"pagination": pagination,
		"data":       data,
	})
}
