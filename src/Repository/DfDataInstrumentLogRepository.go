package Repository

import (
	"gorm.io/gorm"
	"data-fetcher-api/config/packages"
	"time"
)

type DfDataInstrumentLog struct {
	Type       string    `json:"type"`
	PlanId     int       `json:"PlanId"`
	Instrument string    `json:"instrument"`
	Time       time.Time `json:"time"`
}

func CreateOrUpdateInstrumentLog(
	db *gorm.DB,
	planId int,
	typeId string,
	instrument string,
	newData DfDataInstrumentLog,
) (*DfDataInstrumentLog, error) {
	instrumentLog := &DfDataInstrumentLog{}

	result := db.
		Where(DfDataInstrumentLog{Type: typeId, PlanId: planId, Instrument: instrument}).
		Assign(newData).
		FirstOrCreate(instrumentLog)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		result := db.Model(DfDataInstrumentLog{}).
			Where(DfDataInstrumentLog{Type: typeId, PlanId: planId, Instrument: instrument}).
			Update("time", newData.Time)
		if result.Error != nil {
			return nil, result.Error
		}
		return &newData, nil
	}

	return instrumentLog, nil
}

func GetLastExportedDataTime(
	instrument string,
	planId int,
	typeId string) (time.Time, error) {
	db := packages.ConnectTimescaleDb()
	defer packages.CloseDatabaseConnection(db)

	var times time.Time

	query := db.Table("df_data_instrument_logs").
		Select("time").Where("type = ? AND instrument = ? AND plan_id = ?",
		typeId, instrument, planId).Limit(1)

	if err := query.Find(&times).Error; err != nil {
		return times, err
	}

	return times, nil
}
