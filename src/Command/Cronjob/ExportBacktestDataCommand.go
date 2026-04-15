package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"data-fetcher-api/config/packages"
	"data-fetcher-api/src/Api"
	"data-fetcher-api/src/Helper"
	"data-fetcher-api/src/Repository"
	"time"
)

type EndpointResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

const (
	FILE_EXTENSION       = ".txt"
	DATA_INTERVAL_TICK   = "Tick"
	DATA_INTERVAL_DAY    = "Day"
	DATA_INTERVAL_MINUTE = "Minute"
	PUBLIC_FOLDER        = "public/data/export"
	TIME_FORMAT          = "20060102 150405"
	DATE_FORMAT          = "20060102"
	TEMP_DIRECTORY       = "public/data/temp"
	ZIP_EXTENSION        = ".zip"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("env file not loaded")
	}
}

func main() {
	packages.SentryInit()
	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()
	err := ExportData()
	if err != nil {
		sentry.CaptureException(err)
	}
}

func ExportData() error {

	db := packages.ConnectTimescaleDb()
	sqlDB, err := db.DB()
	if err != nil {
		sentry.CaptureException(err)
	} else {
		sqlDB.SetConnMaxLifetime(4 * time.Hour)
	}

	defer packages.CloseDatabaseConnection(db)
	tickDataErr := ExportTickData(DATA_INTERVAL_TICK, db)
	if tickDataErr != nil {
		return tickDataErr
	}
	minuteDataErr := ExportMinuteData(DATA_INTERVAL_MINUTE, db)
	if minuteDataErr != nil {
		return minuteDataErr
	}
	dayDataErr := ExportDailyData(DATA_INTERVAL_DAY, db)
	if dayDataErr != nil {
		return dayDataErr
	}

	return nil
}

func ExportTickData(typeId string, db *gorm.DB) error {
	years := getYearsList()
	for _, year := range years {
		instruments, err := Repository.GetInstruments(typeId, year, db)
		if err != nil {
			sentry.CaptureException(err)
		}
		if len(instruments) == 0 {
			continue
		}
		for _, instrument := range instruments {
			var index = 1
			for month := time.January; month <= time.December; month++ {
				historicalData, dataErr := Repository.GetOtherSourceData(db, typeId, instrument.Instrument, year, month)
				if dataErr != nil {
					sentry.CaptureException(dataErr)
				}
				if len(historicalData) == 0 {
					historicalData, _ = Repository.GetTradingViewSourceData(db, typeId, instrument.Instrument, year, month)
					if len(historicalData) == 0 {
						index++
						continue
					}
				} else {
					Repository.UpdateTradingViewSourceData(db, typeId, instrument.Instrument,
						historicalData[1].Time, historicalData[len(historicalData)-1].Time)
				}
				planIdStr := strconv.Itoa(instrument.PlanID)
				yearStr := strconv.Itoa(year)
				folderPath := filepath.Join(PUBLIC_FOLDER, planIdStr+"/"+DATA_INTERVAL_TICK+"/"+instrument.Instrument+"/"+yearStr)
				if err := Helper.CreateFolderIfNotExists(filepath.Join(PUBLIC_FOLDER, planIdStr)); err != nil {
					return fmt.Errorf("failed to create folder: %w", err)
				}
				if err := Helper.CreateFolderIfNotExists(folderPath); err != nil {
					return fmt.Errorf("failed to create folder: %w", err)
				}

				fileName := strconv.Itoa(index) + FILE_EXTENSION
				file, err := os.Create(PUBLIC_FOLDER + "/" + planIdStr + "/" + DATA_INTERVAL_TICK + "/" + instrument.Instrument + "/" + yearStr + "/" + fileName)

				if err != nil {
					return fmt.Errorf("failed to create file %s: %w", file, err)
				}

				defer func(file *os.File) {
					err := file.Close()
					if err != nil {
						fmt.Print(err.Error())
					}
				}(file)

				writer := csv.NewWriter(file)
				writer.Comma = ';'

				for _, value := range historicalData {
					row := []string{
						value.Time.Format(TIME_FORMAT),
						fmt.Sprintf("%.10g", value.Open),
						fmt.Sprintf("%.10g", value.Volume),
					}
					writer.Write(row)
				}
				writer.Flush()
				time.Sleep(5 * time.Second)
				uploadErr := uploadData(instrument.PlanID)
				if uploadErr != nil {
					return uploadErr
				}

				lastData := historicalData[len(historicalData)-1]
				instrumentLog := Repository.DfDataInstrumentLog{
					Type:       typeId,
					Instrument: instrument.Instrument,
					PlanId:     instrument.PlanID,
					Time:       lastData.Time,
				}

				Repository.CreateOrUpdateInstrumentLog(db, instrument.PlanID, typeId, instrument.Instrument, instrumentLog)
				log.Println("Tick historicalData exported")
				index++
			}
		}
	}
	return nil
}

func ExportMinuteData(typeId string, db *gorm.DB) error {
	years := getYearsList()
	for _, year := range years {
		instruments, err := Repository.GetInstruments(typeId, year, db)
		if err != nil {
			sentry.CaptureException(err)
		}
		if len(instruments) == 0 {
			continue
		}
		for _, instrument := range instruments {
			var index = 1
			for month := time.January; month <= time.December; month++ {
				historicalData, dataErr := Repository.GetOtherSourceData(db, typeId, instrument.Instrument, year, month)
				if dataErr != nil {
					sentry.CaptureException(dataErr)
				}
				if len(historicalData) == 0 {
					historicalData, _ = Repository.GetTradingViewSourceData(db, typeId, instrument.Instrument, year, month)
					if len(historicalData) == 0 {
						index++
						continue
					}
				} else {
					Repository.UpdateTradingViewSourceData(db, typeId, instrument.Instrument, historicalData[1].Time,
						historicalData[len(historicalData)-1].Time)
				}
				planIdStr := strconv.Itoa(instrument.PlanID)
				yearStr := strconv.Itoa(year)
				folderPath := filepath.Join(PUBLIC_FOLDER, planIdStr+"/"+DATA_INTERVAL_MINUTE+"/"+instrument.Instrument+"/"+yearStr)
				if err := Helper.CreateFolderIfNotExists(filepath.Join(PUBLIC_FOLDER, planIdStr)); err != nil {
					return fmt.Errorf("failed to create folder: %w", err)
				}
				if err := Helper.CreateFolderIfNotExists(folderPath); err != nil {
					return fmt.Errorf("failed to create folder: %w", err)
				}

				fileName := strconv.Itoa(index) + FILE_EXTENSION
				file, err := os.Create(PUBLIC_FOLDER + "/" + planIdStr + "/" + DATA_INTERVAL_MINUTE + "/" + instrument.Instrument + "/" + yearStr + "/" + fileName)

				if err != nil {
					return fmt.Errorf("failed to create file %s: %w", file, err)
				}

				defer func(file *os.File) {
					err := file.Close()
					if err != nil {
						fmt.Print(err.Error())
					}
				}(file)

				writer := csv.NewWriter(file)
				writer.Comma = ';'

				for _, value := range historicalData {
					row := []string{
						value.Time.Format(TIME_FORMAT),
						fmt.Sprintf("%.10g", value.Open),
						fmt.Sprintf("%.10g", value.High),
						fmt.Sprintf("%.10g", value.Low),
						fmt.Sprintf("%.10g", value.Close),
						fmt.Sprintf("%.10g", value.Volume),
					}
					writer.Write(row)
				}
				writer.Flush()
				time.Sleep(5 * time.Second)
				uploadErr := uploadData(instrument.PlanID)
				if err != nil {
					return uploadErr
				}

				lastData := historicalData[len(historicalData)-1]
				instrumentLog := Repository.DfDataInstrumentLog{
					Type:       typeId,
					Instrument: instrument.Instrument,
					PlanId:     instrument.PlanID,
					Time:       lastData.Time,
				}

				Repository.CreateOrUpdateInstrumentLog(db, instrument.PlanID, typeId, instrument.Instrument, instrumentLog)
				log.Println("Minute historicalData exported")
				index++
			}
		}
	}
	return nil
}

func ExportDailyData(typeId string, db *gorm.DB) error {
	years := getYearsList()
	for _, year := range years {
		instruments, _ := Repository.GetInstruments(typeId, year, db)
		if len(instruments) == 0 {
			continue
		}
		for _, instrument := range instruments {
			var index = 1
			for month := time.January; month <= time.December; month++ {
				historicalData, dataErr := Repository.GetOtherSourceData(db, typeId, instrument.Instrument, year, month)
				if dataErr != nil {
					sentry.CaptureException(dataErr)
				}
				if len(historicalData) == 0 {
					historicalData, _ = Repository.GetTradingViewSourceData(db, typeId, instrument.Instrument, year, month)
					if len(historicalData) == 0 {
						index++
						continue
					}
				} else {
					Repository.UpdateTradingViewSourceData(db, typeId, instrument.Instrument, historicalData[1].Time,
						historicalData[len(historicalData)-1].Time)
				}
				planIdStr := strconv.Itoa(instrument.PlanID)
				yearStr := strconv.Itoa(year)
				folderPath := filepath.Join(PUBLIC_FOLDER, planIdStr+"/"+DATA_INTERVAL_DAY+"/"+instrument.Instrument+"/"+yearStr)

				if err := Helper.CreateFolderIfNotExists(filepath.Join(PUBLIC_FOLDER, planIdStr)); err != nil {
					return fmt.Errorf("failed to create folder: %w", err)
				}
				if err := Helper.CreateFolderIfNotExists(folderPath); err != nil {
					return fmt.Errorf("failed to create folder: %w", err)
				}

				fileName := strconv.Itoa(index) + FILE_EXTENSION
				file, err := os.Create(PUBLIC_FOLDER + "/" + planIdStr + "/" + DATA_INTERVAL_DAY + "/" + instrument.Instrument + "/" + yearStr + "/" + fileName)

				if err != nil {
					return fmt.Errorf("failed to create file %s: %w", file, err)
				}

				defer func(file *os.File) {
					err := file.Close()
					if err != nil {
						fmt.Print(err.Error())
					}
				}(file)
				writer := csv.NewWriter(file)
				writer.Comma = ';'

				for _, value := range historicalData {
					row := []string{
						value.Time.Format(DATE_FORMAT),
						fmt.Sprintf("%.10g", value.Open),
						fmt.Sprintf("%.10g", value.High),
						fmt.Sprintf("%.10g", value.Low),
						fmt.Sprintf("%.10g", value.Close),
						fmt.Sprintf("%.10g", value.Volume),
					}
					writer.Write(row)
				}
				writer.Flush()
				time.Sleep(5 * time.Second)
				uploadErr := uploadData(instrument.PlanID)
				if err != nil {
					return uploadErr
				}
				lastData := historicalData[len(historicalData)-1]
				instrumentLog := Repository.DfDataInstrumentLog{
					Type:       typeId,
					Instrument: instrument.Instrument,
					PlanId:     instrument.PlanID,
					Time:       lastData.Time,
				}
				Repository.CreateOrUpdateInstrumentLog(db, instrument.PlanID, typeId, instrument.Instrument, instrumentLog)
				log.Println("Day historicalData exported")
				index++
			}
		}
	}
	return nil
}

func getYearsList() []int {
	currentYear := time.Now().Year()
	startYear := currentYear - 80

	yearsList := make([]int, 0)
	for year := currentYear; year >= startYear; year-- {
		yearsList = append(yearsList, year)
	}
	for i, j := 0, len(yearsList)-1; i < j; i, j = i+1, j-1 {
		yearsList[i], yearsList[j] = yearsList[j], yearsList[i]
	}
	return yearsList
}

func uploadData(id int) error {
	sourceLink := GetSource(id)
	if sourceLink != "" {
		planIdStr := strconv.Itoa(id)
		plansId, _ := Helper.GetFolderNames(PUBLIC_FOLDER)
		marketId, _ := Api.GetFolderIDFromURL(sourceLink)
		client, err := Api.CreateClient()
		if err != nil {
			return err
		}
		var instrumentDriveLink string
		var intervalTypeDriveLink string
		var instrumentDriveExist bool
		var intervalTypeDriveExist bool
		var yearDriveExist bool
		var yearDriveLink string
		var fileExist bool
		var fileLink string
		for _, planId := range plansId {
			if planId == planIdStr {
				dataTypes, err := Helper.GetFolderNames(PUBLIC_FOLDER + "/" + planId)
				if err != nil || len(dataTypes) == 0 {
					return err
				}
				for _, dataType := range dataTypes {
					intervalTypeDriveLink, intervalTypeDriveExist, err = Api.CheckFolderExists(client, marketId, dataType)
					if err != nil {
						return err

					}
					if !intervalTypeDriveExist {
						intervalTypeDriveLink, err = Api.CreateFolder(client, marketId, dataType)
						if err != nil {
							return err
						}
					}
					instruments, err := Helper.GetFolderNames(PUBLIC_FOLDER + "/" + planId + "/" + dataType)
					if err != nil || len(instruments) == 0 {
						return err
					}
					for _, instrument := range instruments {
						intervalTypeFolderId, err := Api.GetFolderIDFromURL(intervalTypeDriveLink)
						if err != nil {
							return err
						}
						instrumentDriveLink, instrumentDriveExist, err = Api.CheckFolderExists(client, intervalTypeFolderId, instrument)
						if err != nil {
							return err
						}
						if !instrumentDriveExist {
							instrumentDriveLink, err = Api.CreateFolder(client, intervalTypeFolderId, instrument)
							if err != nil {
								return err
							}
						}
						yearList, _ := Helper.GetFolderNames(PUBLIC_FOLDER + "/" + planId + "/" + dataType + "/" + instrument)
						if err != nil || len(yearList) == 0 {
							return err
						}
						for _, year := range yearList {
							instrumentFolderId, err := Api.GetFolderIDFromURL(instrumentDriveLink)
							if err != nil {
								return err
							}
							yearDriveLink, yearDriveExist, err = Api.CheckFolderExists(client, instrumentFolderId, year)
							if err != nil {
								return err
							}
							if !yearDriveExist {
								yearDriveLink, err = Api.CreateFolder(client, instrumentFolderId, year)
								if err != nil {
									return err
								}
							}
							fileNames, err := Helper.GetFilenamesFromDirectory(PUBLIC_FOLDER + "/" + planId + "/" + dataType + "/" + instrument + "/" + year)
							if err != nil {
								return err
							}
							for _, fileName := range fileNames {
								yearFolderId, err := Api.GetFolderIDFromURL(yearDriveLink)
								if err != nil {
									return err
								}
								fileNameInZiP := Helper.GetNameFromFileName(fileName) + ZIP_EXTENSION
								fileLink, fileExist, err = Api.CheckFileExists(client, yearFolderId, fileNameInZiP)
								if err != nil {
									return err
								}
								exportedfilePath := PUBLIC_FOLDER + "/" + planId + "/" + dataType + "/" + instrument + "/" + year + "/" + fileName
								zipFilePath := PUBLIC_FOLDER + "/" + planId + "/" + dataType + "/" + instrument + "/" + year + "/" + fileNameInZiP
								if !fileExist {

									err := Helper.CreateZip(exportedfilePath, zipFilePath)
									if err != nil {
										return err
									}
									url, err := Api.UploadFile(client, yearFolderId, zipFilePath, fileNameInZiP)
									if err != nil {
										return err
									}
									if url != "" {
										os.Remove(exportedfilePath)
										os.Remove(zipFilePath)
									}
								} else {
									fileId, _ := Api.GetFolderIDFromURL(fileLink)
									tempDownloadDirectory := TEMP_DIRECTORY + "/" + dataType + "/" + instrument
									Helper.CreateFolderIfNotExists(tempDownloadDirectory)
									downloadedFilePath := tempDownloadDirectory + "/" + fileNameInZiP
									err := Api.DownloadFile(client, fileId, downloadedFilePath)
									if err != nil {
										return err
									}
									extractedDirectory := tempDownloadDirectory + "/" + Helper.GetNameFromFileName(fileName)
									err = Helper.ExtractZip(downloadedFilePath, extractedDirectory)
									if err != nil {
										return err
									}
									extractedFilePath := tempDownloadDirectory + "/" + Helper.GetNameFromFileName(fileName) + "/" + fileName
									mergeFilePath := exportedfilePath
									err = Helper.MergeFiles(extractedFilePath, exportedfilePath, mergeFilePath)
									if err != nil {
										return err
									}
									err = Helper.CreateZip(mergeFilePath, zipFilePath)
									if err != nil {
										return err
									}
									url, err := Api.UploadFile(client, yearFolderId, zipFilePath, fileNameInZiP)
									if err != nil {
										return err
									}

									if url != "" {
										Api.DeleteFile(client, fileId)
										os.Remove(mergeFilePath)
									}
								}
							}

						}

					}
				}
				os.RemoveAll(PUBLIC_FOLDER + "/" + planId)
				os.RemoveAll(TEMP_DIRECTORY)
			}
		}
	}
	return nil
}

func GetSource(planId int) string {

	config, _ := Helper.GetParameter()
	productApi := config.Parameters.DfProductApiUri
	response, err := http.Get(fmt.Sprintf("%s/guest/products/%d", productApi, planId))
	if err != nil {
		log.Println("Failed to make the API call:", err)
		return ""
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Failed to read the response body:", err)
		return ""
	}

	defer response.Body.Close()

	var endpointResponse EndpointResponse
	err = json.Unmarshal(body, &endpointResponse)
	if err != nil {
		log.Println("Failed to parse API response:", err)
		return ""
	}

	return endpointResponse.Data
}
