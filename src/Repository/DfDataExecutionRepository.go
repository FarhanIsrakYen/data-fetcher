package Repository

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
	"data-fetcher-api/config/packages"
	"time"
)

const SIGNAL_EXIT = "exit"
const SIGNAL_SELL = "sell"
const SIGNAL_ENTER = "enter"
const SIGNAL_BUY = "buy"
const SIGNAL_RERVRSE = "reverse"
const SIGNAL_CLOSE = "close"
const SIGNAL_CLOSE_SELL = "sell(close)"
const SIGNAL_CLOSE_BUY = "buy(close)"
const QUANTITY = 1
const POSITION_LONG = "long"
const POSITION_FLAT = "flat"
const POSITION_SHORT = "short"

const DEFAULT_LIMIT = 100
const DEFAULT_ORDERBY = "time"
const DEFAULT_ORDER = "DESC"
const DEFAULT_PAGE = 1

type DfDataExecution struct {
	TemplateId int       `json:"templateId"`
	Signals    string    `json:"signals"`
	Position   string    `json:"position"`
	Instrument string    `json:"instrument"`
	Quantity   int       `json:"quantity"`
	Price      *float32  `json:"price" gorm:"type:double precision"`
	Time       time.Time `json:"time"`
}

func CreateExecution(
	db *gorm.DB,
	templateId int,
	signals string,
	position string,
	instrument string,
	quantity int,
	price *float32,
	time time.Time) (*DfDataExecution, error) {

	execution := &DfDataExecution{
		TemplateId: templateId,
		Signals:    signals,
		Position:   position,
		Instrument: instrument,
		Quantity:   quantity,
		Price:      price,
		Time:       time,
	}
	result := db.Create(execution)

	if result.Error != nil {
		return nil, result.Error
	}
	return execution, nil
}

func GetExecutions(
	limit int,
	orderBy string,
	order string,
	page int,
	filters map[string][]string) ([]DfDataExecution, Pagination, error) {

	if limit == 0 {
		limit = DEFAULT_LIMIT
	}
	if len(orderBy) == 0 {
		orderBy = DEFAULT_ORDERBY
	}
	if len(order) == 0 {
		order = DEFAULT_ORDER
	}
	if page == 0 {
		page = DEFAULT_PAGE
	}

	filtersExact, filtersRelated := GetFilters(filters)

	db := packages.ConnectTimescaleDb()

	var executions []DfDataExecution
	var totalItems int64
	offset := (page - 1) * limit

	err := db.Where(filtersExact).
		Where(filtersRelated).
		Offset(offset).
		Limit(limit).
		Order(orderBy + " " + order).
		Find(&executions).Error

	db.Model(&DfDataExecution{}).Where(filtersExact).
		Where(filtersRelated).Count(&totalItems)
	packages.CloseDatabaseConnection(db)

	pagination := GetPagination(page, limit, totalItems)
	if err != nil {
		return nil, pagination, err
	}
	return executions, pagination, nil
}

func GetExecutionsTemplateId(
	db *gorm.DB,
	startTime time.Time,
	endTime time.Time) ([]int, error) {

	var templateIds []int
	result := db.Find(&DfDataExecution{}).
		Where("time >= ? AND time <= ? AND price IS NOT NULL", startTime, endTime).
		Distinct("template_id").
		Pluck("template_id", &templateIds)
	if result.Error != nil {
		return nil, result.Error
	}

	return templateIds, nil
}

func GetLastDayExecutions(
	db *gorm.DB,
	templateId int,
	startTime time.Time,
	endTime time.Time,
) ([]DfDataExecution, error) {

	var executions []DfDataExecution
	offset := 0
	batchSize := 100000

	for {
		var batchExecutions []DfDataExecution

		result := db.Where("time >= ? AND time <= ? AND template_id = ? AND price IS NOT NULL",
			startTime, endTime, templateId).
			Order("time ASC").
			Limit(batchSize).
			Offset(offset).
			Find(&batchExecutions)
		if result.Error != nil {
			return nil, result.Error
		}
		if len(batchExecutions) == 0 {
			break
		}
		executions = append(executions, batchExecutions...)

		if len(batchExecutions) < batchSize {
			break
		}

		offset += batchSize
	}

	return executions, nil
}

func GetExecutionWithoutPrice(db *gorm.DB) ([]DfDataExecution, error) {

	offset := 0
	batchSize := 10000

	var executions []DfDataExecution

	for {
		var batchExecutions []DfDataExecution

		result := db.Where("price IS NULL").
			Limit(batchSize).
			Offset(offset).
			Find(&batchExecutions)

		if result.Error != nil {
			return nil, result.Error
		}
		if len(batchExecutions) == 0 {
			break
		}
		executions = append(executions, batchExecutions...)

		if len(batchExecutions) < batchSize {
			break
		}

		offset += batchSize
	}

	return executions, nil
}

func UpdateExecution(
	db *gorm.DB,
	templateId int,
	signals string,
	position string,
	instrument string,
	price *float32,
	time time.Time,
) (*DfDataExecution, error) {

	execution := &DfDataExecution{
		Price: price,
	}
	result := db.Where("template_id = ? AND instrument = ? AND signals = ? AND position = ? AND time = ?",
		templateId, instrument, signals, position, time).
		Updates(execution)

	if result.Error != nil {
		return nil, result.Error
	}
	return execution, nil
}

func CreateMultipleExecutions(executions []*DfDataExecution) error {
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)

	batchSize := 1000

	for i := 0; i < len(executions); i += batchSize {
		end := i + batchSize
		if end > len(executions) {
			end = len(executions)
		}
		batch := executions[i:end]

		var valueStrings []string
		var valueArgs []interface{}

		for _, execution := range batch {
			valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?)")
			valueArgs = append(valueArgs, execution.TemplateId, execution.Signals, execution.Position, execution.Instrument, execution.Quantity, execution.Price, execution.Time)
		}

		query := fmt.Sprintf("INSERT INTO public.df_data_executions (template_id, signals, position, instrument, quantity, price, time) VALUES %s",
			strings.Join(valueStrings, ","))
		result := db.Exec(query, valueArgs...)

		if result.Error != nil {
			return fmt.Errorf("failed to insert executions: %v", result.Error)
		}
	}

	return nil
}

func DeleteMultipleExecutionsByTemplateIDs(templateIDs []int) error {
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)

	query := fmt.Sprintf("DELETE FROM public.df_data_executions WHERE template_id IN (%s)",
		strings.Trim(strings.Join(strings.Fields(fmt.Sprint(templateIDs)), ","), "[]"))

	result := db.Exec(query)

	return result.Error
}
func GetStrategiesProfits(
	db *gorm.DB,
	templateId int,
	stoploss float32,
	startTime time.Time,
	endTime time.Time,
) float32 {
	executions, _ := GetLastDayExecutions(db, templateId, startTime, endTime)
	profit := float32(0.0)
	enterPrice := float32(0.0)
	lastEnterPrice := float32(0.0)
	exitPrice := float32(0.0)
	lastExitPrice := float32(0.0)
	for _, execution := range executions {
		price := execution.Price

		if execution.Signals == SIGNAL_EXIT ||
			execution.Signals == SIGNAL_SELL {

			lastExitPrice = *price * float32(execution.Quantity)
			currentProfit := lastExitPrice - lastEnterPrice

			if currentProfit <= -stoploss {
				break
			}
			exitPrice += lastExitPrice

			lastEnterPrice = 0
			lastExitPrice = 0
		} else {
			enterPrice += *price * float32(execution.Quantity)
			lastEnterPrice += *price * float32(execution.Quantity)
			lastExitPrice = 0
		}
	}

	profit = exitPrice - enterPrice
	return profit
}

func GetExecutionsByTemplateId(templateId int,
	limit int,
	orderBy string,
	order string,
	page int,
	filters map[string][]string) ([]DfDataExecution, Pagination, error) {

	if limit == 0 {
		limit = DEFAULT_LIMIT
	}
	if len(orderBy) == 0 {
		orderBy = DEFAULT_ORDERBY
	}
	if len(order) == 0 {
		order = DEFAULT_ORDER
	}
	if page == 0 {
		page = DEFAULT_PAGE
	}

	filtersExact, filtersRelated := GetFilters(filters)

	db := packages.ConnectTimescaleDb()
	var totalItems int64
	offset := (page - 1) * limit

	var executions []DfDataExecution
	err := db.Where("template_id = ?", templateId).
		Where(filtersExact).
		Where(filtersRelated).
		Offset(offset).
		Limit(limit).
		Order(orderBy + " " + order).
		Find(&executions).Error

	db.Model(&DfDataExecution{}).
		Where("template_id = ?", templateId).
		Where(filtersExact).Where(filtersRelated).
		Count(&totalItems)

	packages.CloseDatabaseConnection(db)

	pagination := GetPagination(page, limit, totalItems)

	if err != nil {
		return nil, pagination, err
	}
	return executions, pagination, nil

}
