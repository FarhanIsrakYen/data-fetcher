package Model

import (
	"math"
	"data-fetcher-api/src/Helper"
	"data-fetcher-api/src/Repository"
)

const (
	SCORE_ONE                              = 1
	SCORE_TWO                              = 2
	SCORE_THREE                            = 3
	SCORE_FOUR                             = 4
	SCORE_FIVE                             = 5
	SCORE_RANGE_ONE                        = 30
	SCORE_RANGE_TWO                        = 50
	SCORE_RANGE_THREE                      = 60
	SCORE_RANGE_FOUR                       = 80
	SCORE_RANGE_FIVE                       = 100
	RELIABLE_SCORE_MAX_DRAWDOWN            = 20
	RELIABLE_SCORE_MAX_TIME_TO_RECOVER     = 15
	RELIABLE_SCORE_LARGEST_LOSING_TRADE    = 15
	RELIABLE_SCORE_CONSEQUENCE_LOSING      = 15
	RELIABLE_SCORE_AVERAGE_LOSING_TRADE    = 15
	RELIABLE_SCORE_LARGEST_WINNING_TRADE   = 10
	RELIABLE_SCORE_NUMBER_OF_TRADE_PER_DAY = 5
	RELIABLE_SCORE_CONSEQUENCE_WINNING     = 5
)

func GetReliableMetricsScore(
	realTimeMetrics map[string]float32,
	backtestingMetrics map[string]float32) map[string]float32 {

	resultMap := make(map[string]float32)
	for key := range realTimeMetrics {
		if _, ok := backtestingMetrics[key]; ok {
			value := reliableMetricsCalculation(realTimeMetrics[key], backtestingMetrics[key])

			resultMap[key] = value
		}
	}
	return resultMap
}

func reliableMetricsCalculation(
	realTimeMetricsValue float32,
	backtestingMetricsValue float32) float32 {

	realTimeMetricsValue = absValue(realTimeMetricsValue)
	backtestingMetricsValue = absValue(backtestingMetricsValue)
	values := []float32{realTimeMetricsValue, backtestingMetricsValue}
	min, max := Helper.GetMinMax(values)

	result := ((max - min) * 100) / max
	if math.IsNaN(float64(result)) {
		result = 0
	}

	return result
}

func GetReliableScore(reliableMetrics map[string]float32) int {
	var result float32
	for key := range reliableMetrics {

		switch key {
		case Repository.PARAMETER_KEY_MAX_DRAWDOWN:
			result += calculateReliableScore(reliableMetrics[key], RELIABLE_SCORE_MAX_DRAWDOWN)
		case Repository.PARAMETER_KEY_MAX_TIME_TO_RECOVER:
			result += calculateReliableScore(reliableMetrics[key], RELIABLE_SCORE_MAX_TIME_TO_RECOVER)

		case Repository.PARAMETER_KEY_LARGEST_LOSING_TRADE:
			result += calculateReliableScore(reliableMetrics[key], RELIABLE_SCORE_LARGEST_LOSING_TRADE)

		case Repository.PARAMETER_KEY_CONSEQUENCE_LOSING:
			result += calculateReliableScore(reliableMetrics[key], RELIABLE_SCORE_CONSEQUENCE_LOSING)

		case Repository.PARAMETER_KEY_AVERAGE_LOSING_TRADE:
			result += calculateReliableScore(reliableMetrics[key], RELIABLE_SCORE_AVERAGE_LOSING_TRADE)

		case Repository.PARAMETER_KEY_AVERAGE_WINNING_TRADE:
			result += calculateReliableScore(reliableMetrics[key], RELIABLE_SCORE_LARGEST_WINNING_TRADE)

		case Repository.PARAMETER_KEY_NUMBER_OF_TRADE_PER_DAY:
			result += calculateReliableScore(reliableMetrics[key], RELIABLE_SCORE_NUMBER_OF_TRADE_PER_DAY)

		case Repository.PARAMETER_KEY_CONSEQUENCE_WINNING:
			result += calculateReliableScore(reliableMetrics[key], RELIABLE_SCORE_CONSEQUENCE_WINNING)
		}
	}
	return getScore(result)
}

func calculateReliableScore(value float32, percentage int) float32 {
	return float32(percentage) / 100 * (100 - value)
}

func getScore(result float32) int {
	var score int
	switch {
	case result <= SCORE_RANGE_ONE:
		score = SCORE_ONE
	case result <= SCORE_RANGE_TWO:
		score = SCORE_TWO
	case result <= SCORE_RANGE_THREE:
		score = SCORE_THREE
	case result <= SCORE_RANGE_FOUR:
		score = SCORE_FOUR
	case result <= SCORE_RANGE_FIVE:
		score = SCORE_FIVE
	}
	return score
}

func absValue(value float32) float32 {
	return float32(math.Abs(float64(value)))
}
