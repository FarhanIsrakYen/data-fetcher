package HistoricalPerformanceTemplateController

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"data-fetcher-api/config/packages"
)

func GetPerformance(ctx *gin.Context) {
	c := packages.CacheDeclare(false)
	idStr := ctx.Param("id")
	templateId, _ := strconv.Atoi(idStr)

	var key = "performances" + "-" + "historical" + "-" + strconv.Itoa(templateId)

	if performances, found := c.Get(key); found {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "List of historical performances",
			"data":    performances,
		})
	} else {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  true,
			"message": "Historical Performance not found!",
			"data":    []interface{}{},
		})
		return
	}
}
