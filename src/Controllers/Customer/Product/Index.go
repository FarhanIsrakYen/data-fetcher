package ProductController

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"data-fetcher-api/src/Api"
	Controller "data-fetcher-api/src/Controllers"
	"data-fetcher-api/src/Helper"
	"data-fetcher-api/src/Lib"
	"data-fetcher-api/src/Repository"
	"time"
)

const (
	USER_SUBSCRIBED_PRODUCT_REDIS_KEY = "SubscribedProducts"
)

type SubscribedProduct struct {
	PlanId     int
	SourceLink string
}
type ProductStatusResponse struct {
	Time       time.Time `json:"time"`
	SourceLink string    `json:"sourceLink"`
}

func GetProductStatus(ctx *gin.Context) {

	planIdStr := ctx.Query("planId")
	planId, _ := strconv.Atoi(planIdStr)
	instrument := ctx.Query("instrument")
	dataType := ctx.Query("type")

	data, _ := Controller.GetUser(ctx)
	strData := strconv.Itoa(data.ID)
	parameter, _ := Helper.GetParameter()
	key := parameter.Parameters.DfSessionPrefix + USER_SUBSCRIBED_PRODUCT_REDIS_KEY + "-" + strData

	value, _ := Lib.GetValue(key)

	if value != "" {
		var products []SubscribedProduct
		if err := json.Unmarshal([]byte(value), &products); err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "PlanId not found",
				"data":    planId,
			})
			return
		}

		var product SubscribedProduct
		for _, subscribedProduct := range products {
			if subscribedProduct.PlanId == planId {
				product = subscribedProduct
			}
		}
		client, err := Api.CreateClient()
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "Instrument not found",
				"data":    instrument,
			})
			return
		}

		if product.PlanId == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "PlanId not found",
				"data":    planId,
			})
			return
		}
		time, _ := Repository.GetLastExportedDataTime(instrument, planId, dataType)

		if time.IsZero() {
			ctx.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "Instrument not found",
				"data":    instrument,
			})
			return
		}
		var instrumentDriveLink string
		var intervalTypeDriveLink string
		var instrumentDriveExist bool
		var intervalTypeDriveExist bool
		marketId, _ := Api.GetFolderIDFromURL(product.SourceLink)
		intervalTypeDriveLink, intervalTypeDriveExist, err = Api.CheckFolderExists(client, marketId, dataType)
		if intervalTypeDriveExist {
			intervalTypeFolderId, _ := Api.GetFolderIDFromURL(intervalTypeDriveLink)
			instrumentDriveLink, instrumentDriveExist, err = Api.CheckFolderExists(client, intervalTypeFolderId, instrument)
			if instrumentDriveExist {
				var data = ProductStatusResponse{
					Time:       time,
					SourceLink: instrumentDriveLink,
				}
				ctx.JSON(http.StatusOK, gin.H{
					"status":  true,
					"message": "Instrument data retrieved successfully",
					"data":    data,
				})
				return
			}
		}
	}
	ctx.JSON(http.StatusNotFound, gin.H{
		"status":  false,
		"message": "Instrument not found",
		"data":    instrument,
	})
}
