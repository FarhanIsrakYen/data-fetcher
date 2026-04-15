package HistoricalExecutionTemplateController

import (
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"data-fetcher-api/src/Repository"
)

func GetExecution(ctx *gin.Context) {

	idStr := ctx.Param("id")
	templateId, _ := strconv.Atoi(idStr)
	limitStr := ctx.Query("limit")
	limit, _ := strconv.Atoi(limitStr)
	orderBy := ctx.Query("orderBy")
	order := ctx.Query("order")
	pageStr := ctx.Query("page")
	page, _ := strconv.Atoi(pageStr)
	filters := ctx.Request.URL.Query()

	executions, pagination, err := Repository.GetHistoricalExecutionsByTemplateId(templateId, limit, orderBy, order, page, filters)

	if err != nil {
		sentry.CaptureException(err)
	}

	if len(executions) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  true,
			"message": "Historical Execution not found!",
			"data":    []interface{}{},
		})
		return
	}

	for i, _ := range executions {
		if executions[i].Signals == Repository.SIGNAL_CLOSE_SELL ||
			executions[i].Signals == Repository.SIGNAL_CLOSE_BUY {
			executions[i].Signals = Repository.SIGNAL_CLOSE
		} else if executions[i].Signals != Repository.SIGNAL_CLOSE_SELL &&
			executions[i].Signals != Repository.SIGNAL_CLOSE_BUY && executions[i].Quantity > 1 {
			executions[i].Signals = Repository.SIGNAL_RERVRSE
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":     true,
		"message":    "List of historical executions",
		"pagination": pagination,
		"data":       executions,
	})
}
