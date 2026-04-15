package Migrations

import (
	"gorm.io/gorm"
)

func CreateHistoricalPerformanceHyperTable(db *gorm.DB) {

	var tableName = "df_data_historical_performances"
	var exists bool
	err := db.Raw(`SELECT EXISTS (SELECT 1 FROM _timescaledb_catalog.hypertable WHERE schema_name = 'public' AND table_name = '` + tableName + `');`).Row().Scan(&exists)

	if err != nil {
		panic(err.Error())
		return
	}

	if !exists {
		if err := db.Exec("SELECT create_hypertable('" + tableName + "', 'time', chunk_time_interval => INTERVAL '1 day');").Error; err != nil {
			panic(err.Error())
			return
		}
		println("Hyper Table : " + tableName + " created successfully")
		return
	}
	println("Already " + tableName + " Hyper Table Exist")
}
