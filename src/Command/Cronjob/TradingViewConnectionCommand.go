package main

import (
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	socket "github.com/verzth/tradingview-scraper/v2"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"data-fetcher-api/config/packages"
	"data-fetcher-api/src/Helper"
	"data-fetcher-api/src/Repository"
	"time"
)

type InstrumentEndpointResponse struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Pagination Pagination   `json:"pagination"`
	Data       []Instrument `json:"data"`
}

type Instrument struct {
	PlanId         int    `json:"planId"`
	InstrumentName string `json:"instrumentName"`
}

type Pagination struct {
	PagesTotal int `json:"pagesTotal"`
}

const (
	INSTRUMENTS_ENDPOINT = "/guest/instruments"
	MIN_DATA_STORE_LIMIT = 500
	REQUEST_TYPE_GET     = "GET"
	MAX_DATA_LIMIT       = 200
	DEAFULT_PAGE         = 1
)

var instruments []Instrument
var addedSymbols = make(map[string]Instrument)
var instrumentValueDict []Repository.DfDataData
var db *gorm.DB
var err error
var tradingviewsocket socket.SocketInterface

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("env file not loaded")
	}
}

func main() {
	packages.SentryInit()
	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()
	instruments, err = getInstruments()
	if err != nil {
		sentry.CaptureException(err)
		return
	}
	if len(instruments) == 0 {
		return
	}
	db = packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)
	go func() {
		Connect()
	}()

	select {}
}

func Connect() {

	tradingviewsocket, err = socket.Connect(getData, handleError)
	if err != nil {
		panic("Error while initializing the tradingview websocket -> " + err.Error())
	}

	for _, instrument := range instruments {
		if _, exists := addedSymbols[instrument.InstrumentName]; !exists {
			tradingviewsocket.AddSymbol(instrument.InstrumentName)
			addedSymbols[instrument.InstrumentName] = Instrument{
				PlanId: instrument.PlanId,
			}
		}
	}
}

func getData(symbol string, data *socket.QuoteData) {
	if _, exists := addedSymbols[symbol]; exists {
		if data.Price != nil && data.Volume != nil {
			instrumentValue := Repository.DfDataData{
				Type:       Repository.DATA_INTERVAL_TICK,
				Instrument: symbol,
				PlanId:     addedSymbols[symbol].PlanId,
				Open:       float32(*data.Price),
				Volume:     float32(*data.Volume),
				Time:       time.Now(),
				IsExported: false,
				Source:     Repository.HISTORICAL_DATA_SOURCE_TRADINGVIEW,
			}
			instrumentValueDict = append(instrumentValueDict, instrumentValue)
		}
	}
	if len(instrumentValueDict) >= MIN_DATA_STORE_LIMIT {
		err = Repository.CreateMultipleData(db, instrumentValueDict)
		if err != nil {
			panic(err)
		}
		instrumentValueDict = []Repository.DfDataData{}

	}
}
func handleError(err error, context string) {
	fmt.Printf("error -> %s\n", err.Error())
	fmt.Printf("context -> %s\n", context)
}

func getInstruments() ([]Instrument, error) {
	var allInstruments []Instrument
	config, _ := Helper.GetParameter()
	billingApi := config.Parameters.CoreMarketApiUri
	url := billingApi + INSTRUMENTS_ENDPOINT
	client := &http.Client{}

	page := DEAFULT_PAGE
	for {
		req, err := http.NewRequest(REQUEST_TYPE_GET, url, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		query := req.URL.Query()
		query.Add("limit", strconv.Itoa(MAX_DATA_LIMIT))
		query.Add("filters[is_supported]", "true")
		query.Add("page", strconv.Itoa(page))
		req.URL.RawQuery = query.Encode()

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error sending request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %v", err)
		}

		var endpointResponse InstrumentEndpointResponse
		err = json.Unmarshal(body, &endpointResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse API response: %v", err)
		}
		allInstruments = append(allInstruments, endpointResponse.Data...)
		if page >= endpointResponse.Pagination.PagesTotal {
			break
		}
		page++
	}
	return allInstruments, nil
}
