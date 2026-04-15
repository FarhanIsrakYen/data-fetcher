package Mq

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"data-fetcher-api/config/packages"
	"data-fetcher-api/src/Helper"
	MqPublishLib "data-fetcher-api/src/Lib/MqPublish"
	"data-fetcher-api/src/Model"
	"data-fetcher-api/src/Repository"
	"time"
)

type Strategy struct {
	TemplateId int
	Instrument string
	Verified   bool `json:"isVerified"`
	Quantity   string
	Stoploss   string
	StartTime  int64
	EndTime    int64
}
type Payload struct {
	Strategies  []Strategy        `json:"strategies"`
	Strategy    ExecutionStrategy `json:"strategy"`
	Time        int64             `json:"time"`
	UserId      int               `json:"userId"`
	TemplateIds []int             `json:"templateIds"`
}
type ExecutionStrategy struct {
	TemplateId      int
	Instrument      string
	Signal          string
	Quantity        int
	Position        string
	Time            string
	CurrentQuantity int
}

const NUMBER_OF_EXECUTIONS = 52
const TIME_FORMAT = "20060102 150405"

func CreateExecutionAndPerformance(payload string) error {

	var message Payload
	if err := json.Unmarshal([]byte(payload), &message); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}
	for i := 0; i < len(message.Strategies); i++ {
		startFrom := message.Time
		randomTimestamp := make([]int64, 0, 52)
		for j := 0; j < NUMBER_OF_EXECUTIONS; j++ {
			nextTimestamp := GetNextTimestamp(startFrom)
			randomTimestamp = append(randomTimestamp, nextTimestamp[rand.Intn(14)])
			startFrom = nextTimestamp[13]
		}
		executions := CreateStrategiesExecutions(randomTimestamp, message.Strategies[i])
		err := Repository.CreateMultipleExecutions(executions)
		if err != nil {
			log.Println(err)
			return err
		}
		createDummyPerformance(randomTimestamp, message.Strategies[i].TemplateId, false)

		if message.Strategies[i].Verified == true {
			t := time.Unix(message.Time, 0)
			oneYearAgo := t.AddDate(-1, 0, 0)
			historicalStartDate := oneYearAgo.Unix()
			historicalRandomTimestamp := make([]int64, 0, 52)
			for j := 0; j < NUMBER_OF_EXECUTIONS; j++ {
				historicalNextTimestamp := GetNextTimestamp(historicalStartDate)
				historicalRandomTimestamp = append(historicalRandomTimestamp, historicalNextTimestamp[rand.Intn(14)])
				historicalStartDate = historicalNextTimestamp[13]
			}
			historicalExecutions := CreateHistoricalStrategiesExecutions(historicalRandomTimestamp, message.Strategies[i])
			historicalErr := Repository.CreateMultipleHistoricalExecutions(historicalExecutions)
			if err != nil {
				log.Println(historicalErr)
				return historicalErr
			}
			createDummyPerformance(historicalRandomTimestamp, message.Strategies[i].TemplateId, true)
		}
	}
	log.Println("Execution and performance Generated")
	return nil
}

func CreateStrategiesExecutions(
	timeStamp []int64,
	strategies Strategy) []*Repository.DfDataExecution {

	executions := make([]*Repository.DfDataExecution, 0, len(timeStamp)*3)
	execution := make([]*Repository.DfDataExecution, 0, 1*3)

	var nextPrice float32
	var price float32

	priceRangeOne := []string{"NAS100USD", "BTCUSD", "US30"}
	priceRangeTwo := []string{"SP500", "SPXUSD", "ETHUSDT"}
	priceRangeThree := []string{"EURJPY", "USDJPY", "DXY", "BCHUSDT", "BNBUSDT"}
	priceRangeFour := []string{"USDTHB", "SOLUSDT", "KHC", "X", "BKR", "WMB", "PFE"}
	priceRangeFive := []string{"USDCAD", "EURUSD", "USDSGD"}

	priceOne := generateRandomNumber(25994.0, 34831.43)
	priceTwo := generateRandomNumber(4219.89, 5000.3453)
	priceThree := generateRandomNumber(180.264, 235.242)
	priceFour := generateRandomNumber(19.79, 35.141)
	priceFive := generateRandomNumber(1.6532394742965698, 1.9932394742965698)

	for i := 0; i < len(timeStamp); i++ {

		switch {
		case Helper.ArrayContains(priceRangeOne, strategies.Instrument):
			price = priceOne

			price1 := price
			price2 := price1 + float32(200.0001)
			price3 := price + float32(1000.4)

			randomIndexOne := generateRandomIntArray(1, 12, 4)
			randomIndexTwo := generateRandomIntArray(10, 30, 3)
			randomIndexThree := generateRandomIntArray(30, 46, 5)

			if randomIndexOne[0] == i || randomIndexOne[1] == i || randomIndexOne[2] == i || randomIndexOne[3] == i {
				price = nextPrice + float32(700.02)
			}
			if randomIndexTwo[0] == i || randomIndexTwo[1] == i || randomIndexTwo[2] == i {
				price = nextPrice + float32(1500.02)
			}
			if randomIndexThree[0] == i || randomIndexThree[1] == i || randomIndexThree[2] == i || randomIndexThree[3] == i ||
				randomIndexThree[4] == i {
				price = nextPrice + generateRandomNumber(200.01, 600)
			}

			execution = createDummyExecutions(timeStamp[i], strategies, price1, price2, price3)
			executions = append(executions, execution...)
		case Helper.ArrayContains(priceRangeTwo, strategies.Instrument):
			price = priceTwo
			price1 := price
			price2 := price1 + float32(50.0001)
			price3 := price + float32(200.4)

			randomIndexOne := generateRandomIntArray(1, 17, 5)
			randomIndexTwo := generateRandomIntArray(10, 30, 4)
			randomIndexThree := generateRandomIntArray(32, 47, 6)

			if randomIndexOne[0] == i || randomIndexOne[1] == i || randomIndexOne[2] == i || randomIndexOne[3] == i ||
				randomIndexOne[4] == i {
				price = nextPrice + float32(50.02)
			}
			if randomIndexTwo[0] == i || randomIndexTwo[1] == i || randomIndexTwo[2] == i || randomIndexTwo[3] == i {
				price = nextPrice + float32(250.003)
			}
			if randomIndexThree[0] == i || randomIndexThree[1] == i || randomIndexThree[2] == i || randomIndexThree[3] == i ||
				randomIndexThree[4] == i || randomIndexThree[5] == i {
				price = nextPrice + generateRandomNumber(180.0634, 220.0454)
			}
			execution = createDummyExecutions(timeStamp[i], strategies, price1, price2, price3)
			executions = append(executions, execution...)

		case Helper.ArrayContains(priceRangeThree, strategies.Instrument):
			price = priceThree
			price1 := price
			price2 := price1 + float32(10.00)
			price3 := price + float32(15.4)

			randomIndexOne := generateRandomIntArray(1, 15, 4)
			randomIndexTwo := generateRandomIntArray(11, 32, 6)
			randomIndexThree := generateRandomIntArray(25, 45, 5)

			if randomIndexOne[0] == i || randomIndexOne[1] == i || randomIndexOne[2] == i || randomIndexOne[3] == i {
				price = nextPrice + float32(5.02)
			}
			if randomIndexTwo[0] == i || randomIndexTwo[1] == i || randomIndexTwo[2] == i || randomIndexTwo[3] == i ||
				randomIndexTwo[4] == i || randomIndexTwo[5] == i {
				price = nextPrice + float32(10.003)
			}
			if randomIndexThree[0] == i || randomIndexThree[1] == i || randomIndexThree[2] == i || randomIndexThree[3] == i ||
				randomIndexThree[4] == i {
				price = nextPrice + generateRandomNumber(3.89, 6.00)
			}
			execution = createDummyExecutions(timeStamp[i], strategies, price1, price2, price3)
			executions = append(executions, execution...)

		case Helper.ArrayContains(priceRangeFour, strategies.Instrument):
			price = priceFour
			price1 := price
			price2 := price1 + float32(2.00)
			price3 := price + float32(5.4)

			randomIndexOne := generateRandomIntArray(5, 15, 4)
			randomIndexTwo := generateRandomIntArray(11, 28, 6)
			randomIndexThree := generateRandomIntArray(25, 45, 5)

			if randomIndexOne[0] == i || randomIndexOne[1] == i || randomIndexOne[2] == i || randomIndexOne[3] == i {
				price = nextPrice + float32(1.02)
			}
			if randomIndexTwo[0] == i || randomIndexTwo[1] == i || randomIndexTwo[2] == i || randomIndexTwo[3] == i ||
				randomIndexTwo[4] == i || randomIndexTwo[5] == i {
				price = nextPrice + float32(3.003)
			}
			if randomIndexThree[0] == i || randomIndexThree[1] == i || randomIndexThree[2] == i || randomIndexThree[3] == i ||
				randomIndexThree[4] == i {
				price = nextPrice + generateRandomNumber(1.00, 2.03)
			}
			execution = createDummyExecutions(timeStamp[i], strategies, price1, price2, price3)
			executions = append(executions, execution...)

		case Helper.ArrayContains(priceRangeFive, strategies.Instrument):
			price = priceFive
			price1 := price
			price2 := price1 + float32(0.0001)
			price3 := price + float32(0.4)

			randomIndexOne := generateRandomIntArray(1, 20, 5)
			randomIndexTwo := generateRandomIntArray(12, 25, 1)
			randomIndexThree := generateRandomIntArray(25, 47, 5)

			if randomIndexOne[0] == i || randomIndexOne[1] == i || randomIndexOne[2] == i || randomIndexOne[3] == i ||
				randomIndexOne[4] == i {
				price = nextPrice + float32(0.05)
			}
			if randomIndexTwo[0] == i {
				price = nextPrice + float32(0.16)
			}
			if randomIndexThree[0] == i || randomIndexThree[1] == i || randomIndexThree[2] == i || randomIndexThree[3] == i ||
				randomIndexThree[4] == i {
				price = nextPrice + float32(0.03)
			}
			execution = createDummyExecutions(timeStamp[i], strategies, price1, price2, price3)
			executions = append(executions, execution...)
		}
		nextPrice = price
		priceOne = nextPrice - generateRandomNumber(350.00, 500.00)
		priceTwo = nextPrice - generateRandomNumber(80.00, 120.00)
		priceThree = nextPrice - generateRandomNumber(2.00, 5.00)
		priceFour = nextPrice - generateRandomNumber(0.5, 1)
		priceFive = nextPrice - float32(0.02)
	}
	return executions
}

func CreateHistoricalStrategiesExecutions(
	timeStamp []int64,
	strategies Strategy) []*Repository.DfDataHistoricalExecution {

	executions := make([]*Repository.DfDataHistoricalExecution, 0, len(timeStamp)*3)
	execution := make([]*Repository.DfDataHistoricalExecution, 0, 1*3)

	var nextPrice float32
	var price float32

	priceRangeOne := []string{"NAS100USD", "BTCUSD", "US30"}
	priceRangeTwo := []string{"SP500", "SPXUSD", "ETHUSDT"}
	priceRangeThree := []string{"EURJPY", "USDJPY", "DXY", "BCHUSDT", "BNBUSDT"}
	priceRangeFour := []string{"USDTHB", "SOLUSDT", "KHC", "X", "BKR", "WMB", "PFE"}
	priceRangeFive := []string{"USDCAD", "EURUSD", "USDSGD"}

	priceOne := generateRandomNumber(25994.0, 34831.43)
	priceTwo := generateRandomNumber(4219.89, 5000.3453)
	priceThree := generateRandomNumber(180.264, 235.242)
	priceFour := generateRandomNumber(19.79, 35.141)
	priceFive := generateRandomNumber(1.6532394742965698, 1.9932394742965698)

	for i := 0; i < len(timeStamp); i++ {

		switch {
		case Helper.ArrayContains(priceRangeOne, strategies.Instrument):
			price = priceOne

			price1 := price
			price2 := price1 + float32(200.0001)
			price3 := price + float32(1000.4)

			randomIndexOne := generateRandomIntArray(1, 12, 4)
			randomIndexTwo := generateRandomIntArray(10, 30, 3)
			randomIndexThree := generateRandomIntArray(30, 46, 5)

			if randomIndexOne[0] == i || randomIndexOne[1] == i || randomIndexOne[2] == i || randomIndexOne[3] == i {
				price = nextPrice + float32(700.02)
			}
			if randomIndexTwo[0] == i || randomIndexTwo[1] == i || randomIndexTwo[2] == i {
				price = nextPrice + float32(1500.02)
			}
			if randomIndexThree[0] == i || randomIndexThree[1] == i || randomIndexThree[2] == i || randomIndexThree[3] == i ||
				randomIndexThree[4] == i {
				price = nextPrice + generateRandomNumber(200.01, 600)
			}

			execution = createDummyHistoricalExecutions(timeStamp[i], strategies, price1, price2, price3)
			executions = append(executions, execution...)
		case Helper.ArrayContains(priceRangeTwo, strategies.Instrument):
			price = priceTwo
			price1 := price
			price2 := price1 + float32(50.0001)
			price3 := price + float32(200.4)

			randomIndexOne := generateRandomIntArray(1, 17, 5)
			randomIndexTwo := generateRandomIntArray(10, 30, 4)
			randomIndexThree := generateRandomIntArray(32, 47, 6)

			if randomIndexOne[0] == i || randomIndexOne[1] == i || randomIndexOne[2] == i || randomIndexOne[3] == i ||
				randomIndexOne[4] == i {
				price = nextPrice + float32(50.02)
			}
			if randomIndexTwo[0] == i || randomIndexTwo[1] == i || randomIndexTwo[2] == i || randomIndexTwo[3] == i {
				price = nextPrice + float32(250.003)
			}
			if randomIndexThree[0] == i || randomIndexThree[1] == i || randomIndexThree[2] == i || randomIndexThree[3] == i ||
				randomIndexThree[4] == i || randomIndexThree[5] == i {
				price = nextPrice + generateRandomNumber(180.0634, 220.0454)
			}
			execution = createDummyHistoricalExecutions(timeStamp[i], strategies, price1, price2, price3)
			executions = append(executions, execution...)

		case Helper.ArrayContains(priceRangeThree, strategies.Instrument):
			price = priceThree
			price1 := price
			price2 := price1 + float32(10.00)
			price3 := price + float32(15.4)

			randomIndexOne := generateRandomIntArray(1, 15, 4)
			randomIndexTwo := generateRandomIntArray(11, 32, 6)
			randomIndexThree := generateRandomIntArray(25, 45, 5)

			if randomIndexOne[0] == i || randomIndexOne[1] == i || randomIndexOne[2] == i || randomIndexOne[3] == i {
				price = nextPrice + float32(5.02)
			}
			if randomIndexTwo[0] == i || randomIndexTwo[1] == i || randomIndexTwo[2] == i || randomIndexTwo[3] == i ||
				randomIndexTwo[4] == i || randomIndexTwo[5] == i {
				price = nextPrice + float32(10.003)
			}
			if randomIndexThree[0] == i || randomIndexThree[1] == i || randomIndexThree[2] == i || randomIndexThree[3] == i ||
				randomIndexThree[4] == i {
				price = nextPrice + generateRandomNumber(3.89, 6.00)
			}
			execution = createDummyHistoricalExecutions(timeStamp[i], strategies, price1, price2, price3)
			executions = append(executions, execution...)

		case Helper.ArrayContains(priceRangeFour, strategies.Instrument):
			price = priceFour
			price1 := price
			price2 := price1 + float32(2.00)
			price3 := price + float32(5.4)

			randomIndexOne := generateRandomIntArray(5, 15, 4)
			randomIndexTwo := generateRandomIntArray(11, 28, 6)
			randomIndexThree := generateRandomIntArray(25, 45, 5)

			if randomIndexOne[0] == i || randomIndexOne[1] == i || randomIndexOne[2] == i || randomIndexOne[3] == i {
				price = nextPrice + float32(1.02)
			}
			if randomIndexTwo[0] == i || randomIndexTwo[1] == i || randomIndexTwo[2] == i || randomIndexTwo[3] == i ||
				randomIndexTwo[4] == i || randomIndexTwo[5] == i {
				price = nextPrice + float32(3.003)
			}
			if randomIndexThree[0] == i || randomIndexThree[1] == i || randomIndexThree[2] == i || randomIndexThree[3] == i ||
				randomIndexThree[4] == i {
				price = nextPrice + generateRandomNumber(1.00, 2.03)
			}
			execution = createDummyHistoricalExecutions(timeStamp[i], strategies, price1, price2, price3)
			executions = append(executions, execution...)

		case Helper.ArrayContains(priceRangeFive, strategies.Instrument):
			price = priceFive
			price1 := price
			price2 := price1 + float32(0.0001)
			price3 := price + float32(0.4)

			randomIndexOne := generateRandomIntArray(1, 20, 5)
			randomIndexTwo := generateRandomIntArray(12, 25, 1)
			randomIndexThree := generateRandomIntArray(25, 47, 5)

			if randomIndexOne[0] == i || randomIndexOne[1] == i || randomIndexOne[2] == i || randomIndexOne[3] == i ||
				randomIndexOne[4] == i {
				price = nextPrice + float32(0.05)
			}
			if randomIndexTwo[0] == i {
				price = nextPrice + float32(0.16)
			}
			if randomIndexThree[0] == i || randomIndexThree[1] == i || randomIndexThree[2] == i || randomIndexThree[3] == i ||
				randomIndexThree[4] == i {
				price = nextPrice + float32(0.03)
			}
			execution = createDummyHistoricalExecutions(timeStamp[i], strategies, price1, price2, price3)
			executions = append(executions, execution...)
		}
		nextPrice = price
		priceOne = nextPrice - generateRandomNumber(350.00, 500.00)
		priceTwo = nextPrice - generateRandomNumber(80.00, 120.00)
		priceThree = nextPrice - generateRandomNumber(2.00, 5.00)
		priceFour = nextPrice - generateRandomNumber(0.5, 1)
		priceFive = nextPrice - float32(0.02)
	}
	return executions
}

func generateRandomNumber(min float32, max float32) float32 {
	rand.Seed(time.Now().UnixNano())

	return min + rand.Float32()*(max-min)
}

func generateRandomIntArray(min int, max int, count int) []int {
	rand.Seed(time.Now().UnixNano())

	randomIntArray := make([]int, count)
	for i := 0; i < count; i++ {
		randomIntArray[i] = rand.Intn(max-min+1) + min
	}

	return randomIntArray
}

func createDummyExecutions(
	timeStamp int64,
	strategies Strategy,
	price1 float32,
	price2 float32,
	price3 float32) []*Repository.DfDataExecution {
	executions := make([]*Repository.DfDataExecution, 0, 1*3)
	execution1 := &Repository.DfDataExecution{
		TemplateId: strategies.TemplateId,
		Signals:    Repository.SIGNAL_BUY,
		Position:   Repository.POSITION_LONG,
		Instrument: strategies.Instrument,
		Quantity:   Repository.QUANTITY,
		Price:      &price1,
		Time:       Helper.TimestampToTime(timeStamp),
	}

	execution2 := &Repository.DfDataExecution{
		TemplateId: strategies.TemplateId,
		Signals:    Repository.SIGNAL_BUY,
		Position:   Repository.POSITION_LONG,
		Instrument: strategies.Instrument,
		Quantity:   Repository.QUANTITY + Repository.QUANTITY,
		Price:      &price2,
		Time:       Helper.TimestampToTime(timeStamp),
	}

	execution3 := &Repository.DfDataExecution{
		TemplateId: strategies.TemplateId,
		Signals:    Repository.SIGNAL_SELL,
		Position:   Repository.POSITION_FLAT,
		Instrument: strategies.Instrument,
		Quantity:   Repository.QUANTITY + Repository.QUANTITY + Repository.QUANTITY,
		Price:      &price3,
		Time:       Helper.TimestampToTime(timeStamp),
	}
	executions = append(executions, execution1, execution2, execution3)
	return executions
}

func createDummyHistoricalExecutions(
	timeStamp int64,
	strategies Strategy,
	price1 float32,
	price2 float32,
	price3 float32) []*Repository.DfDataHistoricalExecution {
	executions := make([]*Repository.DfDataHistoricalExecution, 0, 1*3)
	execution1 := &Repository.DfDataHistoricalExecution{
		TemplateId: strategies.TemplateId,
		Signals:    Repository.SIGNAL_BUY,
		Instrument: strategies.Instrument,
		Quantity:   Repository.QUANTITY,
		Price:      &price1,
		Time:       Helper.TimestampToTime(timeStamp),
	}

	execution2 := &Repository.DfDataHistoricalExecution{
		TemplateId: strategies.TemplateId,
		Signals:    Repository.SIGNAL_BUY,
		Instrument: strategies.Instrument,
		Quantity:   Repository.QUANTITY + Repository.QUANTITY,
		Price:      &price2,
		Time:       Helper.TimestampToTime(timeStamp),
	}

	execution3 := &Repository.DfDataHistoricalExecution{
		TemplateId: strategies.TemplateId,
		Signals:    Repository.SIGNAL_SELL,
		Instrument: strategies.Instrument,
		Quantity:   Repository.QUANTITY + Repository.QUANTITY + Repository.QUANTITY,
		Price:      &price3,
		Time:       Helper.TimestampToTime(timeStamp),
	}
	executions = append(executions, execution1, execution2, execution3)
	return executions
}

func createDummyPerformance(
	timeStamp []int64,
	templateId int,
	isHistorical bool) {
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)

	for i := 0; i < len(timeStamp); i++ {
		firstSecond, lastSecond := Helper.TimestampToFirstAndLastMinuteOfDay(timeStamp[i])
		if isHistorical {
			Model.CalculateHistoricalPerformance(db, templateId, firstSecond, lastSecond)
		} else {
			Model.CalculatePerformance(db, templateId, firstSecond, lastSecond)
		}
	}
}

func GetNextTimestamp(startFrom int64) []int64 {
	var timestamps []int64

	for i := 0; i < 14; i++ {
		nextDay := startFrom + int64((i+1)*86400)
		timestamps = append(timestamps, nextDay)
	}

	return timestamps
}

func RemoveExecutionAndPerformance(payload string) error {
	var message Payload
	if err := json.Unmarshal([]byte(payload), &message); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}
	executionErr := Repository.DeleteMultipleExecutionsByTemplateIDs(message.TemplateIds)
	if executionErr != nil {
		fmt.Println(executionErr)
		return nil
	}
	historicalExecutionErr := Repository.DeleteMultipleHistoricalExecutionsByTemplateIDs(message.TemplateIds)
	if historicalExecutionErr != nil {
		fmt.Println(historicalExecutionErr)
		return nil
	}
	performanceErr := Repository.DeleteMultiplePerformanceByTemplateIDs(message.TemplateIds)
	if performanceErr != nil {
		fmt.Println(performanceErr)
		return nil
	}
	historicalPerformanceErr := Repository.DeleteMultipleHistoricalPerformanceByTemplateIDs(message.TemplateIds)
	if historicalPerformanceErr != nil {
		fmt.Println(historicalPerformanceErr)
		return nil
	}
	fmt.Println("Execution and performance removed successfully")
	return nil
}

func GetUserProfit(payload string) error {

	var message Payload
	if err := json.Unmarshal([]byte(payload), &message); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}
	profit := float32(0.0)
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)
	for i := 0; i < len(message.Strategies); i++ {
		quantity, err := strconv.ParseFloat(message.Strategies[i].Quantity, 32)
		stoploss, err := strconv.ParseFloat(message.Strategies[i].Stoploss, 32)
		if err != nil {
			return fmt.Errorf("failed to parse strategy profit: %v", err)
		}

		startTime := Helper.TimestampToTime(message.Strategies[i].StartTime)
		endTime := Helper.TimestampToTime(message.Strategies[i].EndTime)
		profit += Repository.GetStrategiesProfits(db, message.Strategies[i].TemplateId, float32(stoploss), startTime, endTime) * float32(quantity)
	}
	topic := "/api/tc/data/profit/create"
	profitMessage := map[string]interface{}{
		"userId": message.UserId,
		"profit": profit,
	}
	jsonData, err := json.Marshal(profitMessage)
	if err != nil {
		fmt.Println("Failed to create user strategy profit")
		return nil
	}
	MqPublishLib.MqPublish(string(jsonData), topic)
	return nil
}

func CreateStrategyExecution(payload string) error {

	var message Payload
	if err := json.Unmarshal([]byte(payload), &message); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	signalTime, _ := time.Parse(TIME_FORMAT, message.Strategy.Time)

	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)

	if message.Strategy.Signal == Repository.SIGNAL_RERVRSE {
		if message.Strategy.CurrentQuantity < 0 {
			message.Strategy.CurrentQuantity *= -2
			message.Strategy.Signal = Repository.SIGNAL_BUY
		} else {
			message.Strategy.CurrentQuantity *= 2
			message.Strategy.Signal = Repository.SIGNAL_SELL
		}
	} else if message.Strategy.Signal == Repository.SIGNAL_CLOSE {
		if message.Strategy.CurrentQuantity < 0 {
			message.Strategy.CurrentQuantity *= -1
			message.Strategy.Signal = Repository.SIGNAL_CLOSE_BUY
		} else {
			message.Strategy.CurrentQuantity *= 1
			message.Strategy.Signal = Repository.SIGNAL_CLOSE_SELL
		}
	} else {
		message.Strategy.CurrentQuantity = Repository.QUANTITY
	}

	tenMinutesAgo := time.Now().Add(-10 * time.Minute)
	if tenMinutesAgo.After(signalTime) {
		_, err := Repository.CreateHistoricalExecution(
			db,
			message.Strategy.TemplateId,
			message.Strategy.Signal,
			message.Strategy.Instrument,
			message.Strategy.CurrentQuantity,
			nil,
			signalTime)

		if err != nil {
			log.Println(err)
			return nil
		}

	} else {
		_, err := Repository.CreateExecution(
			db,
			message.Strategy.TemplateId,
			message.Strategy.Signal,
			message.Strategy.Position,
			message.Strategy.Instrument,
			message.Strategy.CurrentQuantity,
			nil,
			signalTime)

		if err != nil {
			log.Println(err)
			return nil
		}
	}
	year, month, day := signalTime.Date()
	startOfSignalDay := time.Date(year, month, day, 0, 0, 0, 0, signalTime.Location())
	Repository.CreatePerformanceLog(db, message.Strategy.TemplateId, startOfSignalDay)
	log.Println("Execution Created Successfully")
	return nil
}
