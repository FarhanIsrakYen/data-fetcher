package Migrations

import (
	"gorm.io/gorm"
)

func CreateDataHyperTable(db *gorm.DB) {
	var tableName = "df_data_data"
	var exists bool
	err := db.Raw(`SELECT EXISTS (SELECT 1 FROM _timescaledb_catalog.hypertable WHERE schema_name = 'public' AND table_name = '` + tableName + `');`).Row().Scan(&exists)

	if err != nil {
		panic(err.Error())
	}

	if !exists {
		db.Exec("SELECT create_hypertable('" + tableName + "', 'time');")
		println("Hyper Table : " + tableName + " created successfully")
		return
	}
	db.Exec("SELECT set_chunk_time_interval('" + tableName + "', INTERVAL '15 day');")
	println("Already " + tableName + " Hyper Table Exist")
}
