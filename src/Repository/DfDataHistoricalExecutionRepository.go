package Repository

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
	"data-fetcher-api/config/packages"
	"time"
)

type DfDataHistoricalExecution struct {
	TemplateId int       `json:"templateId"`
	Signals    string    `json:"signals"`
	Instrument string    `json:"instrument"`
	Quantity   int       `json:"quantity"`
	Price      *float32  `json:"price" gorm:"type:double precision"`
	Time       time.Time `json:"time"`
}

func CreateHistoricalExecution(
	db *gorm.DB,
	templateId int,
	signals string,
	instrument string,
	quantity int,
	price *float32,
	time time.Time) (*DfDataHistoricalExecution, error) {

	execution := &DfDataHistoricalExecution{
		TemplateId: templateId,
		Signals:    signals,
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

func GetHistoricalExecutions(
	limit int,
	orderBy string,
	order string,
	page int,
	filters map[string][]string) ([]DfDataHistoricalExecution, Pagination, error) {

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

	var executions []DfDataHistoricalExecution
	var totalItems int64
	offset := (page - 1) * limit

	err := db.Where(filtersExact).
		Where(filtersRelated).
		Offset(offset).
		Limit(limit).
		Order(orderBy + " " + order).
		Find(&executions).Error

	db.Model(&DfDataHistoricalExecution{}).Where(filtersExact).
		Where(filtersRelated).Count(&totalItems)
	packages.CloseDatabaseConnection(db)

	pagination := GetPagination(page, limit, totalItems)
	if err != nil {
		return nil, pagination, err
	}
	return executions, pagination, nil
}

func GetHistoricalExecutionsTemplateId(
	db *gorm.DB,
	startTime time.Time,
	endTime time.Time) ([]int, error) {

	var templateIds []int
	result := db.Find(&DfDataHistoricalExecution{}).
		Where("time >= ? AND time <= ? AND price IS NOT NULL", startTime, endTime).
		Distinct("template_id").
		Pluck("template_id", &templateIds)
	if result.Error != nil {
		return nil, result.Error
	}
	return templateIds, nil
}

func GetLastDayHistoricalExecutions(
	db *gorm.DB,
	templateId int,
	startTime time.Time,
	endTime time.Time,
) ([]DfDataHistoricalExecution, error) {

	var executions []DfDataHistoricalExecution
	offset := 0
	batchSize := 100000

	for {
		var batchExecutions []DfDataHistoricalExecution

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

func GetHistoricalExecutionWithoutPrice(db *gorm.DB) ([]DfDataHistoricalExecution, error) {

	offset := 0
	batchSize := 10000

	var executions []DfDataHistoricalExecution

	for {
		var batchExecutions []DfDataHistoricalExecution

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

func UpdateHistoricalExecution(
	db *gorm.DB,
	templateId int,
	signals string,
	instrument string,
	price *float32,
	time time.Time,
) (*DfDataHistoricalExecution, error) {

	execution := &DfDataHistoricalExecution{
		Price: price,
	}
	result := db.Where("template_id = ? AND instrument = ? AND signals = ? AND time = ?",
		templateId, instrument, signals, time).
		Updates(execution)

	if result.Error != nil {
		return nil, result.Error
	}
	return execution, nil
}

func GetHistoricalExecutionsByTemplateId(templateId int,
	limit int,
	orderBy string,
	order string,
	page int,
	filters map[string][]string) ([]DfDataHistoricalExecution, Pagination, error) {

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

	var executions []DfDataHistoricalExecution

	err := db.Where("template_id = ?", templateId).
		Where(filtersExact).
		Where(filtersRelated).
		Offset(offset).
		Limit(limit).
		Order(orderBy + " " + order).
		Find(&executions).Error

	db.Model(&DfDataHistoricalExecution{}).
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

func CreateMultipleHistoricalExecutions(executions []*DfDataHistoricalExecution) error {
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
			valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")
			valueArgs = append(valueArgs, execution.TemplateId, execution.Signals, execution.Instrument, execution.Quantity, execution.Price, execution.Time)
		}

		query := fmt.Sprintf("INSERT INTO public.df_data_historical_executions (template_id, signals, instrument, quantity, price, time) VALUES %s",
			strings.Join(valueStrings, ","))
		result := db.Exec(query, valueArgs...)

		if result.Error != nil {
			return fmt.Errorf("failed to insert historical executions: %v", result.Error)
		}
	}

	return nil
}

func DeleteMultipleHistoricalExecutionsByTemplateIDs(templateIDs []int) error {
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)

	query := fmt.Sprintf("DELETE FROM public.df_data_historical_executions WHERE template_id IN (%s)",
		strings.Trim(strings.Join(strings.Fields(fmt.Sprint(templateIDs)), ","), "[]"))

	result := db.Exec(query)

	return result.Error
}
