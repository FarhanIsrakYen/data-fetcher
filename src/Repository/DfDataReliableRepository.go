package Repository

import (
	"data-fetcher-api/config/packages"
	"time"
)

type DfDataReliable struct {
	TemplateId     int       `json:"templateId"`
	Type           string    `json:"type"`
	ParameterKey   string    `json:"parameterKey"`
	ParameterValue float32   `json:"parameterValue"`
	Time           time.Time `json:"time"`
}

type ReliableData struct {
	ParameterKey   string  `json:"parameterKey"`
	ParameterValue float32 `json:"parameterValue"`
}

const (
	PARAMETER_KEY_MAX_DRAWDOWN              = "max drawdown"
	PARAMETER_KEY_MAX_TIME_TO_RECOVER       = "max time to recover"
	PARAMETER_KEY_CONSEQUENCE_WINNING       = "consequence winning"
	PARAMETER_KEY_CONSEQUENCE_LOSING        = "consequence losing"
	PARAMETER_KEY_AVERAGE_WINNING_TRADE     = "average winning trade"
	PARAMETER_KEY_AVERAGE_LOSING_TRADE      = "average losing trade"
	PARAMETER_KEY_NUMBER_OF_TRADE_PER_DAY   = "number of trades per day"
	PARAMETER_KEY_AVERAGE_PROFIT_PERCENTAGE = "avg. profit percentage"
	PARAMETER_KEY_LARGEST_WINNING_TRADE     = "largest winning trade"
	PARAMETER_KEY_LARGEST_LOSING_TRADE      = "largest losing trade"
	TYPE_REALTIME                           = "realtime"
	TYPE_BACKTESTING                        = "backtesting"
)

func GetReliableDataByTemplateId(
	templateId int,
	dataType string) (map[string]float32, error) {
	var data []ReliableData
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)
	err := db.Model(DfDataReliable{}).
		Where(DfDataReliable{
			TemplateId: templateId,
			Type:       dataType,
		}).Find(&data).Error

	if err != nil {
		return nil, err
	}

	resultMap := make(map[string]float32)
	for _, value := range data {
		resultMap[value.ParameterKey] = value.ParameterValue
	}

	return resultMap, nil

}
