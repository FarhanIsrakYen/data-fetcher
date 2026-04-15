package Repository

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
	"data-fetcher-api/config/packages"
	"time"
)

type DfDataHistoricalPerformance struct {
	TemplateId    int       `json:"templateId"`
	Profit        float32   `json:"profit" gorm:"type:double precision"`
	MaxDrawdown   float32   `json:"maxDrawdown" gorm:"type:double precision"`
	Trades        int       `json:"trades"`
	WinningTrades int       `json:"winningTrades"`
	LosingTrades  int       `json:"losingTrades"`
	Time          time.Time `json:"time"`
}

func CreateHistoricalPerformance(
	db *gorm.DB,
	templateId int,
	profit float32,
	maxDrawDown float32,
	trades int,
	winningTrades int,
	losingTrades int,
	time time.Time) (*DfDataHistoricalPerformance, error) {

	performance := &DfDataHistoricalPerformance{
		TemplateId:    templateId,
		Profit:        profit,
		MaxDrawdown:   maxDrawDown,
		Trades:        trades,
		WinningTrades: winningTrades,
		LosingTrades:  losingTrades,
		Time:          time,
	}
	result := db.Create(performance)
	if result.Error != nil {
		return nil, result.Error
	}
	return performance, nil
}

func GetHistoricalPerformances(
	limit int,
	orderBy string,
	order string,
	page int,
	filters map[string][]string) ([]DfDataHistoricalPerformance, Pagination, error) {

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

	var performances []DfDataHistoricalPerformance
	var totalItems int64
	offset := (page - 1) * limit

	err := db.Where(filtersRelated).
		Where(filtersExact).
		Offset(offset).
		Limit(limit).
		Order(orderBy + " " + order).
		Find(&performances).Error

	db.Model(&DfDataHistoricalPerformance{}).Where(filtersExact).
		Where(filtersRelated).Count(&totalItems)
	packages.CloseDatabaseConnection(db)

	pagination := GetPagination(page, limit, totalItems)

	if err != nil {
		return nil, pagination, err
	}
	return performances, pagination, nil
}

func GetHistoricalPerformanceByTemplateId(
	templateId int,
	order string,
	db *gorm.DB) ([]DfDataHistoricalPerformance, float64, float64, error) {

	var performances []DfDataHistoricalPerformance

	err := db.Where("template_id = ?", templateId).
		Order(DEFAULT_ORDERBY + " " + order).
		Find(&performances).Error

	var avgProfit, avgMaxDrawDown float64
	avgError := float64(0)

	err = db.Model(&DfDataHistoricalPerformance{}).
		Where("template_id = ?", templateId).
		Select("AVG(profit) as avg_profit, AVG(max_drawdown) as avg_max_drawdown").
		Row().
		Scan(&avgProfit, &avgMaxDrawDown)

	if err != nil {
		return nil, avgError, avgError, err
	}
	return performances, avgProfit, avgMaxDrawDown, nil

}

func GetHistoricalPerformancesTemplateId(
	db *gorm.DB) ([]int, error) {

	var templateIds []int
	result := db.Find(&DfDataHistoricalPerformance{}).
		Distinct("template_id").
		Pluck("template_id", &templateIds)
	if result.Error != nil {
		return nil, result.Error
	}

	return templateIds, nil
}

func DeleteMultipleHistoricalPerformanceByTemplateIDs(templateIDs []int) error {
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)

	query := fmt.Sprintf("DELETE FROM public.df_data_historical_performances WHERE template_id IN (%s)",
		strings.Trim(strings.Join(strings.Fields(fmt.Sprint(templateIDs)), ","), "[]"))

	result := db.Exec(query)

	return result.Error
}

func GetLatestAndOldestHistoricalProfit(templateId int) (float32, float32, error) {
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)
	var latestProfit float32
	var oldestProfit float32

	err := db.Table("df_data_historical_performances").
		Where("template_id = ?", templateId).
		Order("time DESC").
		Limit(1).
		Select("profit").
		Scan(&latestProfit).Error
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get latest profit: %w", err)
	}

	err = db.Table("df_data_historical_performances").
		Where("template_id = ?", templateId).
		Order("time ASC").
		Limit(1).
		Select("profit").
		Scan(&oldestProfit).Error
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get oldest profit: %w", err)
	}

	return latestProfit, oldestProfit, nil
}

func HistoricalPerformanceExists(templateID int) bool {
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)
	var count int64

	db.Model(&DfDataHistoricalPerformance{}).Where("template_id = ?", templateID).Count(&count)
	return count > 0
}
