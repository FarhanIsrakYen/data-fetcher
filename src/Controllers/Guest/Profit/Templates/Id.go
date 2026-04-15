package ProfitTemplatesController

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"data-fetcher-api/src/Model"
	"data-fetcher-api/src/Repository"
)

func GetEstimatedProfit(ctx *gin.Context) {
	idStr := ctx.Param("id")
	templateId, _ := strconv.Atoi(idStr)
	timerangeStr := ctx.Query("timerange")
	timerange, _ := strconv.Atoi(timerangeStr)
	isRealTimePerformanceExist := Repository.RealTimePerformanceExists(templateId)

	if isRealTimePerformanceExist {
		latestProfit, oldProfit, _ := Repository.GetLatestAndOldestRealtimeProfit(templateId)
		estimationsProfit := Model.EstimationProfits(latestProfit, oldProfit, templateId, timerange)

		ctx.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "Simulated profit retrieved successfully.",
			"data":    estimationsProfit,
		})
		return
	}
	isHistoricalPerformanceExist := Repository.HistoricalPerformanceExists(templateId)
	if isHistoricalPerformanceExist {
		latestProfit, oldProfit, _ := Repository.GetLatestAndOldestHistoricalProfit(templateId)
		estimationsProfit := Model.EstimationProfits(latestProfit, oldProfit, templateId, timerange)
		ctx.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "Simulated profit retrieved successfully.",
			"data":    estimationsProfit,
		})
		return
	}

	ctx.JSON(http.StatusNotFound, gin.H{
		"status":  false,
		"message": "Trader has not started trading yet",
		"data":    templateId,
	})
	return
}

func GetEstimatedProfitEligibleStrategies(ctx *gin.Context) {
	templatesId := ctx.Query("templatesId")
	templates := strings.Split(templatesId, ",")

	var templatesIntArr []int
	for _, s := range templates {
		num, err := strconv.Atoi(s)
		if err != nil {
			fmt.Printf("Error converting string to int: %v", err)
			return
		}
		templatesIntArr = append(templatesIntArr, num)
	}

	fmt.Println(templatesIntArr)
	templateIDs := Repository.ProfitSimulationEligibleStrategies(templatesIntArr)
	sort.Ints(templateIDs)

	ctx.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Eligible strategies for simulation",
		"data":    templateIDs,
	})
}
