package InstrumentController

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"data-fetcher-api/src/Helper"
)

func GetInstrumentPrice(ctx *gin.Context) {

	data, _ := Helper.GetRealTimeInstrumentPrice(ctx.Param("instrumentName"))
	ctx.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Instrument price retrieved successfully.",
		"data":    data,
	})
}
