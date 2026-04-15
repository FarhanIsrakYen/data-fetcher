package Repository

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"

	"data-fetcher-api/config/packages"
)

type DfDataPerformance struct {
	TemplateId    int       `json:"templateId"`
	Profit        float32   `json:"profit" gorm:"type:double precision"`
	MaxDrawdown   float32   `json:"maxDrawdown" gorm:"type:double precision"`
	Trades        int       `json:"trades"`
	WinningTrades int       `json:"winningTrades"`
	LosingTrades  int       `json:"losingTrades"`
	Time          time.Time `json:"time"`
}

type DfTopPerformance struct {
	TemplateId int `json:"templateId"`
}

const Minimum_Percentage = 0.0

func CreatePerformance(
	db *gorm.DB,
	templateId int,
	profit float32,
	maxDrawDown float32,
	trades int,
	winningTrades int,
	losingTrades int,
	time time.Time) (*DfDataPerformance, error) {

	performance := &DfDataPerformance{
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

func GetPerformances(
	limit int,
	orderBy string,
	order string,
	page int,
	filters map[string][]string) ([]DfDataPerformance, Pagination, error) {

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

	var performances []DfDataPerformance
	var totalItems int64
	offset := (page - 1) * limit

	err := db.Where(filtersRelated).
		Where(filtersExact).
		Offset(offset).
		Limit(limit).
		Order(orderBy + " " + order).
		Find(&performances).Error

	db.Model(&DfDataPerformance{}).Where(filtersExact).
		Where(filtersRelated).Count(&totalItems)
	packages.CloseDatabaseConnection(db)

	pagination := GetPagination(page, limit, totalItems)

	if err != nil {
		return nil, pagination, err
	}
	return performances, pagination, nil
}

func GetPerformanceByTemplateId(
	templateId int,
	order string,
	db *gorm.DB) ([]DfDataPerformance, float64, float64, error) {

	var performances []DfDataPerformance

	err := db.Where("template_id = ?", templateId).
		Order(DEFAULT_ORDERBY + " " + order).
		Find(&performances).Error

	var avgProfit, avgMaxDrawDown float64
	avgError := float64(0)

	err = db.Model(&DfDataPerformance{}).
		Where("template_id = ?", templateId).
		Select("AVG(profit) as avg_profit, AVG(max_drawdown) as avg_max_drawdown").
		Row().
		Scan(&avgProfit, &avgMaxDrawDown)

	if err != nil {
		return nil, avgError, avgError, err
	}
	return performances, avgProfit, avgMaxDrawDown, nil

}

func DeleteMultiplePerformanceByTemplateIDs(templateIDs []int) error {
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)

	query := fmt.Sprintf("DELETE FROM public.df_data_performances WHERE template_id IN (%s)",
		strings.Trim(strings.Join(strings.Fields(fmt.Sprint(templateIDs)), ","), "[]"))

	result := db.Exec(query)

	return result.Error
}

func GetPerformancesTemplateId(
	db *gorm.DB) ([]int, error) {

	var templateIds []int
	result := db.Find(&DfDataPerformance{}).
		Distinct("template_id").
		Pluck("template_id", &templateIds)
	if result.Error != nil {
		return nil, result.Error
	}

	return templateIds, nil
}

func GetTopStrategiesId() ([]DfTopPerformance, error) {
	db := packages.ConnectTimescaleDb()

	var performances []DfTopPerformance

	err := db.Table("df_data_performances").
		Select("template_id, MAX(profit) AS max_profit").
		Group("template_id").
		Order("max_profit DESC").
		Find(&performances).Error

	packages.CloseDatabaseConnection(db)

	if err != nil {
		return nil, err
	}
	return performances, nil
}

func GetLatestAndOldestRealtimeProfit(templateId int) (float32, float32, error) {
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)
	var latestProfit, oldestProfit float32

	if err := db.Table("df_data_performances").
		Where("template_id = ?", templateId).
		Order("time DESC").
		Limit(1).
		Pluck("profit", &latestProfit).Error; err != nil {
		return 0, 0, fmt.Errorf("failed to get latest profit: %w", err)
	}

	if err := db.Table("df_data_performances").
		Where("template_id = ?", templateId).
		Order("time ASC").
		Limit(1).
		Pluck("profit", &oldestProfit).Error; err != nil {
		return 0, 0, fmt.Errorf("failed to get oldest profit: %w", err)
	}

	return latestProfit, oldestProfit, nil
}

func RealTimePerformanceExists(templateID int) bool {
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)
	var count int64

	db.Model(&DfDataPerformance{}).Where("template_id = ?", templateID).Count(&count)
	return count > 0
}

func ProfitSimulationEligibleStrategies(templateIDs []int) []int {
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)
	var resultsRealTime, resultsHistorical []int

	err := db.Table("df_data_performances").
		Select("template_id").
		Where("template_id IN (?)", templateIDs).
		Group("template_id").
		Having("count(*) >= ?", 30).
		Pluck("template_id", &resultsRealTime).Error
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	err = db.Table("df_data_historical_performances").
		Select("template_id").
		Where("template_id IN (?)", templateIDs).
		Group("template_id").
		Having("count(*) >= ?", 30).
		Pluck("template_id", &resultsHistorical).Error
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	results := append(resultsRealTime, resultsHistorical...)
	encountered := map[int]bool{}
	var result []int

	for _, v := range results {
		if !encountered[v] {
			encountered[v] = true
			result = append(result, v)
		}
	}
	return result
}
