package Helper

import (
	"fmt"
	"github.com/verzth/tradingview-scraper/v2"
	"time"
)

func GetRealTimeInstrumentPrice(instrumentName string) (*float64, error) {
	tradingViewSocket := make(chan *tradingview.QuoteData)
	errorCh := make(chan error)

	socket, err := tradingview.Connect(
		func(symbol string, value *tradingview.QuoteData) {
			if symbol == instrumentName {
				tradingViewSocket <- value
			}
		},
		func(err error, context string) {
			errorCh <- fmt.Errorf("%s: %v", context, err)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error connecting to TradingView socket: %v", err)
	}
	defer socket.Close()

	err = socket.AddSymbol(instrumentName)
	if err != nil {
		return nil, fmt.Errorf("error adding symbol: %v", err)
	}

	select {
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout: no data received for instrument %s", instrumentName)
	case err := <-errorCh:
		return nil, err
	case instrumentValue := <-tradingViewSocket:
		return instrumentValue.Price, nil
	}
}
