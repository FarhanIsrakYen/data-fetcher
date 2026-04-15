package Model

import (
	"gorm.io/gorm"
	"math/rand"
	"data-fetcher-api/src/Helper"
	"data-fetcher-api/src/Repository"
	"time"
)

const (
	DEFAULT_TIMERANGE      = 6
	PROFIT_DIVISOR         = 180
	MAX_PERCENTAGE         = 100
	DECREMENT_MINiMUM_RATE = 1
	DECREMENT_MAXIMUM_RATE = 3
)

type EstimationProfit struct {
	TemplateId int       `json:"templateId"`
	Profit     float32   `json:"profit"`
	Time       time.Time `json:"time"`
}

func CalculatePerformance(
	db *gorm.DB,
	templateId int,
	previousDayFirstMin time.Time,
	previousDayLastMin time.Time) *Repository.DfDataPerformance {

	lastDayExecutions, _ := Repository.GetLastDayExecutions(db, templateId, previousDayFirstMin, previousDayLastMin)

	profitPercentage := float32(0.0)
	maxDrawDown := float32(0.0)
	enterPrice := float32(0.0)
	exitPrice := float32(0.0)
	trades := 0
	currentTradePrice := float32(0.0)
	winningTrades := 0
	losingTrades := 0
	var prices []*float32

	for _, execution := range lastDayExecutions {
		price := execution.Price
		prices = append(prices, price)

		if execution.Signals == Repository.SIGNAL_EXIT ||
			execution.Signals == Repository.SIGNAL_SELL ||
			execution.Signals == Repository.SIGNAL_CLOSE_SELL {
			exitPrice += *price * float32(execution.Quantity)
			trades++

			if currentTradePrice < (*price * float32(execution.Quantity)) {
				winningTrades++
			} else {
				losingTrades++
			}
			currentTradePrice = 0
		} else {
			currentTradePrice += *price * float32(execution.Quantity)
			enterPrice += *price * float32(execution.Quantity)
		}
	}
	if enterPrice != 0 {
		profitPercentage = ((exitPrice - enterPrice) * 100) / enterPrice
	} else {
		profitPercentage = Repository.Minimum_Percentage
	}
	if len(prices) == 0 {
		maxDrawDown = Repository.Minimum_Percentage
	} else {
		min, max := Helper.GetMinMax(prices)
		maxDrawDown = ((min - max) * 100) / max
	}

	currentTime := previousDayLastMin
	nextDay := currentTime.AddDate(0, 0, 1)
	times := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, nextDay.Location())

	performance, _ := Repository.CreatePerformance(
		db,
		templateId,
		profitPercentage,
		maxDrawDown,
		trades,
		winningTrades,
		losingTrades,
		times)
	return performance
}

func CalculateHistoricalPerformance(
	db *gorm.DB,
	templateId int,
	previousDayFirstMin time.Time,
	previousDayLastMin time.Time) *Repository.DfDataHistoricalPerformance {

	lastDayExecutions, _ := Repository.GetLastDayHistoricalExecutions(db, templateId, previousDayFirstMin, previousDayLastMin)

	profitPercentage := float32(0.0)
	maxDrawDown := float32(0.0)
	enterPrice := float32(0.0)
	exitPrice := float32(0.0)
	trades := 0
	currentTradePrice := float32(0.0)
	winningTrades := 0
	losingTrades := 0
	var prices []*float32

	for _, execution := range lastDayExecutions {
		price := execution.Price
		prices = append(prices, price)

		if execution.Signals == Repository.SIGNAL_EXIT ||
			execution.Signals == Repository.SIGNAL_SELL ||
			execution.Signals == Repository.SIGNAL_CLOSE_SELL {
			exitPrice += *price * float32(execution.Quantity)
			trades++

			if currentTradePrice < (*price * float32(execution.Quantity)) {
				winningTrades++
			} else {
				losingTrades++
			}
			currentTradePrice = 0
		} else {
			currentTradePrice += *price * float32(execution.Quantity)
			enterPrice += *price * float32(execution.Quantity)
		}
	}
	if enterPrice != 0 {
		profitPercentage = ((exitPrice - enterPrice) * 100) / enterPrice
	} else {
		profitPercentage = Repository.Minimum_Percentage
	}
	if len(prices) == 0 {
		maxDrawDown = Repository.Minimum_Percentage
	} else {
		min, max := Helper.GetMinMax(prices)
		maxDrawDown = ((min - max) * 100) / max
	}

	currentTime := previousDayLastMin
	nextDay := currentTime.AddDate(0, 0, 1)
	times := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, nextDay.Location())

	performance, _ := Repository.CreateHistoricalPerformance(
		db,
		templateId,
		profitPercentage,
		maxDrawDown,
		trades,
		winningTrades,
		losingTrades,
		times)
	return performance
}

func EstimationProfits(
	latestProfit float32,
	oldProfit float32,
	templateId int,
	timerange int) []EstimationProfit {

	var estimationProfit float32
	var profit float32
	if timerange == 0 {
		timerange = DEFAULT_TIMERANGE
	}
	profitDiff := latestProfit - oldProfit
	value := profitDiff / PROFIT_DIVISOR
	totalDecrement := float32(timerange) * Helper.GenerateRandomNumber(DECREMENT_MINiMUM_RATE, DECREMENT_MAXIMUM_RATE)
	upcomingTimerangeDays := Helper.GetUpcomingDays(timerange)
	randomIndices := make([]int, int(totalDecrement))
	for i := 0; i < int(totalDecrement); i++ {
		randomIndices[i] = rand.Intn(len(upcomingTimerangeDays))
	}
	var estimations []EstimationProfit
	for i := 0; i < len(upcomingTimerangeDays); i++ {

		if Helper.IntArrayContains(randomIndices, i) {
			profit -= value
		} else {
			profit += value
		}
		estimationProfit = latestProfit + profit
		if estimationProfit > MAX_PERCENTAGE {
			estimationProfit = MAX_PERCENTAGE
		}
		estimations = append(estimations, EstimationProfit{
			TemplateId: templateId,
			Time:       upcomingTimerangeDays[i],
			Profit:     estimationProfit,
		})
	}
	return estimations
}
