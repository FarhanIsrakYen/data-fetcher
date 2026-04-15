package ReliableScoreController

import (
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"data-fetcher-api/src/Model"
	"data-fetcher-api/src/Repository"
)

type ReliableData struct {
	RealTimeMetrics      map[string]float32 `json:"realTimeMetrics"`
	BacktestingMetrics   map[string]float32 `json:"backtestingMetrics"`
	ReliableMetricsScore map[string]float32 `json:"reliableMetricsScore"`
	Score                int                `json:"score"`
}

func GetReliableScore(ctx *gin.Context) {
	idStr := ctx.Param("id")
	templateId, _ := strconv.Atoi(idStr)
	var errRealTime, errBacktesting error
	realTimeMetrics, errRealTime := Repository.GetReliableDataByTemplateId(templateId, Repository.TYPE_REALTIME)
	backtestingMetrics, errBacktesting := Repository.GetReliableDataByTemplateId(templateId, Repository.TYPE_BACKTESTING)

	if errRealTime != nil {
		sentry.CaptureException(errRealTime)
	}
	if errBacktesting != nil {
		sentry.CaptureException(errBacktesting)
	}
	if len(realTimeMetrics) == 0 || len(backtestingMetrics) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  true,
			"message": "Reliable score not found!",
			"data":    []interface{}{},
		})
		return
	}

	reliableScore := Model.GetReliableMetricsScore(realTimeMetrics, backtestingMetrics)
	score := Model.GetReliableScore(reliableScore)

	ctx.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Reliable score retrieved successfully!",
		"data": ReliableData{
			BacktestingMetrics:   backtestingMetrics,
			RealTimeMetrics:      realTimeMetrics,
			ReliableMetricsScore: reliableScore,
			Score:                score,
		},
	})
}
