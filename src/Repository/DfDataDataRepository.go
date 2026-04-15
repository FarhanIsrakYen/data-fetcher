package Repository

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

type DfDataData struct {
	Type       string    `json:"type"`
	PlanId     int       `json:"PlanId"`
	Instrument string    `json:"instrument"`
	Open       float32   `json:"open" gorm:"type:double precision"`
	High       float32   `json:"high" gorm:"type:double precision"`
	Low        float32   `json:"low" gorm:"type:double precision"`
	Close      float32   `json:"close" gorm:"type:double precision"`
	Volume     float32   `json:"volume" gorm:"type:double precision"`
	Time       time.Time `json:"time"`
	IsExported bool      `json:"isExported" gorm:"default:false"`
	Source     string    `json:"source" gorm:"default:proofinvest"`
}
type Instruments struct {
	Instrument string `json:"instrument"`
	PlanID     int    `json:"plan_id"`
}
type ExportData struct {
	Type   string    `json:"type"`
	Open   float32   `json:"open" gorm:"type:double precision"`
	High   float32   `json:"high" gorm:"type:double precision"`
	Low    float32   `json:"low" gorm:"type:double precision"`
	Close  float32   `json:"close" gorm:"type:double precision"`
	Volume float32   `json:"volume" gorm:"type:double precision"`
	Time   time.Time `json:"time"`
}

const DATA_INTERVAL_TICK = "Tick"
const HISTORICAL_DATA_SOURCE_TRADINGVIEW = "tradingview"

func GetTradingViewSourceData(db *gorm.DB, types string, instrument string, year int, month time.Month) ([]ExportData, error) {
	var batchSize = 50000
	var maxFetch = 300000

	var data []ExportData
	startTime := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endTime := startTime.AddDate(0, 1, -1).Add(time.Hour*23 + time.Minute*59 + time.Second*59 + time.Nanosecond)

	var fetchedCount int
	for fetchedCount < maxFetch {

		remainingFetch := maxFetch - fetchedCount
		if remainingFetch < batchSize {
			batchSize = remainingFetch
		}
		query := db.Table("df_data_data").Where("time >= ? AND time < ? AND type = ? AND instrument = ? AND is_exported = ? AND source = ?",
			startTime, endTime, types, instrument, false, HISTORICAL_DATA_SOURCE_TRADINGVIEW).Limit(batchSize)

		var batchData []ExportData
		if err := query.Find(&batchData).Error; err != nil {
			return nil, err
		}
		if len(batchData) == 0 {
			break
		}

		data = append(data, batchData...)

		if err := db.Table("df_data_data").Where("time >= ? AND time < ? AND type = ? AND instrument = ? AND is_exported = ? AND source = ?",
			startTime, endTime, types, instrument, false, HISTORICAL_DATA_SOURCE_TRADINGVIEW).Limit(batchSize).Update("is_exported", true).Error; err != nil {
			return nil, err
		}
		fetchedCount += len(batchData)
	}
	return data, nil
}

func GetNearestData(
	db *gorm.DB,
	instrument string,
	time time.Time) (*DfDataData, error) {

	var data DfDataData

	result := db.Where("instrument = ? AND time = ?", instrument, time).
		First(&data)

	if result.Error != nil {

		result = db.Where("instrument = ? AND time < ?", instrument, time).
			Order("time DESC").Limit(1).First(&data)

		if result.Error != nil {

			result = db.Where("instrument = ? AND time > ?", instrument, time).
				Order("time ASC").Limit(1).First(&data)

			if result.Error != nil {
				return nil, result.Error
			}
		}
	}
	return &data, nil
}

func GetInstruments(types string, year int, db *gorm.DB) ([]Instruments, error) {

	startTime := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(year, time.December, 31, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	var instruments []Instruments
	result := db.Raw(`
		SELECT DISTINCT instrument, plan_id
		FROM public.df_data_data
		WHERE is_exported = false AND type = ? AND time >= ? AND time < ?`,
		types, startTime, endTime).Scan(&instruments)
	if result.Error != nil {
		return nil, result.Error
	}

	return instruments, nil
}

func CreateMultipleData(db *gorm.DB, realTimeData []DfDataData) error {
	var valueStrings []string
	var valueArgs []interface{}

	for _, realTimeData := range realTimeData {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, realTimeData.Type, realTimeData.PlanId, realTimeData.Instrument, realTimeData.Open, realTimeData.High, realTimeData.Low, realTimeData.Close, realTimeData.Volume, realTimeData.Time, realTimeData.IsExported, realTimeData.Source)
	}

	query := fmt.Sprintf("INSERT INTO public.df_data_data (type, plan_id, instrument, open, high, low, close, volume, time, is_exported,source) VALUES %s", strings.Join(valueStrings, ","))
	result := db.Exec(query, valueArgs...)

	if result.Error != nil {
		return fmt.Errorf("failed to insert data: %v", result.Error)
	}

	return nil
}

func GetOtherSourceData(
	db *gorm.DB,
	types string,
	instrument string,
	year int,
	month time.Month) ([]ExportData, error) {
	var batchSize = 50000
	var maxFetch = 300000

	var data []ExportData
	startTime := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endTime := startTime.AddDate(0, 1, -1).Add(time.Hour*23 + time.Minute*59 + time.Second*59 + time.Nanosecond)

	var fetchedCount int
	for fetchedCount < maxFetch {

		remainingFetch := maxFetch - fetchedCount
		if remainingFetch < batchSize {
			batchSize = remainingFetch
		}
		query := db.Table("df_data_data").Where("time >= ? AND time < ? AND type = ? AND instrument = ? AND is_exported = ? AND source != ?",
			startTime, endTime, types, instrument, false, HISTORICAL_DATA_SOURCE_TRADINGVIEW).Limit(batchSize)

		var batchData []ExportData
		if err := query.Find(&batchData).Error; err != nil {
			return nil, err
		}
		if len(batchData) == 0 {
			break
		}
		data = append(data, batchData...)

		if err := db.Table("df_data_data").Where("time >= ? AND time < ? AND type = ? AND instrument = ? AND is_exported = ?  AND source != ?",
			startTime, endTime, types, instrument, false, HISTORICAL_DATA_SOURCE_TRADINGVIEW).Limit(batchSize).Update("is_exported", true).Error; err != nil {
			return nil, err
		}
		fetchedCount += len(batchData)
	}
	return data, nil
}

func UpdateTradingViewSourceData(
	db *gorm.DB,
	types string,
	instrument string,
	startDay time.Time,
	endDay time.Time) error {
	startTime := time.Date(startDay.Year(), startDay.Month(), startDay.Day(), 0, 0, 0, 0, time.UTC)
	endTime := time.Date(endDay.Year(), endDay.Month(), endDay.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)

	var totalData int64
	db.Model(&DfDataData{}).
		Where("time >= ? AND time < ? AND type = ? AND instrument = ? AND is_exported = ? AND source = ?", startTime, endTime, types, instrument, false, HISTORICAL_DATA_SOURCE_TRADINGVIEW).
		Count(&totalData)
	var batchSize = 50000
	if totalData > 0 {
		if int(totalData) < batchSize {
			batchSize = int(totalData)
		}
		var fetchedCount int
		for fetchedCount < int(totalData) {
			if err := db.Table("df_data_data").Where("time >= ? AND time < ? AND type = ? AND instrument = ? AND is_exported = ?  AND source = ?",
				startTime, endTime, types, instrument, false, HISTORICAL_DATA_SOURCE_TRADINGVIEW).Limit(batchSize).Update("is_exported", true).Error; err != nil {
				return nil
			}
			fetchedCount += batchSize
		}
	}
	return nil
}
