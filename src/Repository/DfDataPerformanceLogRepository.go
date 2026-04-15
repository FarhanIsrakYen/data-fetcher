package Repository

import (
	"gorm.io/gorm"
	"time"
)

type DfDataPerformanceLog struct {
	TemplateId  int       `json:"templateId"`
	Time        time.Time `json:"time"`
	IsGenerated bool      `json:"IsGenerated" gorm:"default:false"`
}

func CreatePerformanceLog(
	db *gorm.DB,
	templateId int,
	time time.Time,
) (*DfDataPerformanceLog, error) {
	performanceLog := &DfDataPerformanceLog{}
	result := db.
		Where(DfDataPerformanceLog{TemplateId: templateId, Time: time}).
		Assign(DfDataPerformanceLog{TemplateId: templateId, Time: time}).
		FirstOrCreate(performanceLog)

	if result.Error != nil {
		return nil, result.Error
	}
	return performanceLog, nil
}

func GetPerformanceLog(
	db *gorm.DB,
	templateId int,
	time time.Time,
) (DfDataPerformanceLog, error) {

	var performanceLog DfDataPerformanceLog
	result := db.
		Where("template_id = ?  AND time = ? ",
			templateId, time).
		Order("time DESC").
		Last(&performanceLog)

	if result.Error != nil {
		return DfDataPerformanceLog{}, result.Error
	}

	return performanceLog, nil
}

func GetOldPerformanceLogData(
	db *gorm.DB,
) (DfDataPerformanceLog, error) {

	var log DfDataPerformanceLog

	query := db.Where("is_generated Is false")
	result := query.Order("time ASC").First(&log)
	if result.Error != nil {
		return DfDataPerformanceLog{}, result.Error
	}

	return log, nil
}

func UpdatePerformanceLog(
	db *gorm.DB,
	time time.Time) error {

	result := db.Model(DfDataPerformanceLog{}).
		Where("time = ?", time).
		Update("is_generated", true)

	if result.Error != nil {
		return result.Error
	}
	return nil
}
