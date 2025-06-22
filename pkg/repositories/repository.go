package repositories

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jame-developer/aeontrac/configuration"
	holidays2 "github.com/jame-developer/aeontrac/pkg/holidays"
	"github.com/jame-developer/aeontrac/pkg/models"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	dataFileName       = "aeon_vault.json"
	BackUpFileNameTmpl = "%d.bak"
	WorkType           = "WORK"
	CompensatoryType   = "COMPENSATORY"
)

// LoadAeonVault loads the time tracking data from the provided folder, using the provided public holdidays configuration, if the data does not exist, it creates a new one.
func LoadAeonVault(folder string, validator *validator.Validate) (models.AeonVault, error) {
	fileName := filepath.Join(folder, dataFileName)
	// Check if the file exists
	if _, err := os.Stat(fileName); err != nil {
		return models.AeonVault{}, err
	}

	file, err := os.Open(fileName)
	if err != nil {
		return models.AeonVault{}, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var currentYear models.AeonVault
	err = json.NewDecoder(file).Decode(&currentYear)
	if err != nil {
		return models.AeonVault{}, err
	}
	validErr := validator.Struct(&currentYear)
	if validErr != nil {
		return models.AeonVault{}, validErr
	}
	for _, aeonDay := range currentYear.Days {
		dayValidErr := validator.Struct(aeonDay)
		if dayValidErr != nil {
			return models.AeonVault{}, dayValidErr
		}
		for _, aeonUnit := range aeonDay.Units {
			unitValidErr := validator.Struct(aeonUnit)
			if unitValidErr != nil {
				return models.AeonVault{}, unitValidErr
			}
		}

	}
	return currentYear, nil

}

// SaveAeonVault saves the time tracking data to the provided folder
func SaveAeonVault(folder string, data models.AeonVault) error {

	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	fileName := filepath.Join(folder, dataFileName)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	_, err = io.WriteString(file, string(jsonData))
	if err != nil {
		return err
	}

	return nil
}

// NewAeonVault creates a new AeonVault instance with the provided year and public holidays configuration
func NewAeonVault(year int, publicHolidaysConfig configuration.PublicHolidaysConfig) (models.AeonVault, error) {
	holidays, err := holidays2.LoadHolidays(publicHolidaysConfig, year)
	if err != nil {
		return models.AeonVault{}, err
	}

	days := map[string]*models.AeonDay{}
	firstDayOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	for d := firstDayOfYear; d.Year() == year; d = d.AddDate(0, 0, 1) {
		dayKey := d.Format(time.DateOnly)
		newDay := NewAoenDay(d)
		if holiday, ok := holidays[dayKey]; ok {
			newDay.PublicHoliday = true
			newDay.PublicHolidayName = holiday.Name
		}

		days[dayKey] = newDay
	}

	return models.AeonVault{
		Days: days,
	}, nil
}

// getIsoDayOfWeek returns the day of the week according to the ISO 8601 standard and the name of the day
func getIsoDayOfWeek(d time.Time) (int, string) {
	day := int(d.Weekday())
	if day == 0 {
		return 7, ""
	}
	return day, d.Weekday().String()
}

// NewAoenDay creates a new AeonDay instance with the provided date
func NewAoenDay(date time.Time) *models.AeonDay {
	_, isoWeekNumber := date.ISOWeek()
	isoDayOfWeekNumber, _ := getIsoDayOfWeek(date)
	isWeekEnd := isoDayOfWeekNumber == 6 || isoDayOfWeekNumber == 7
	return &models.AeonDay{
		IsoWeekNumber: isoWeekNumber,
		IsoWeekDay:    isoDayOfWeekNumber,
		WeekEnd:       isWeekEnd,
		Units:         map[uuid.UUID]models.AeonUnit{},
	}
}

// NewAeonUnit creates a new AeonUnit instance with the provided parameters
func NewAeonUnit(startDateTime, stopDateTime *time.Time, comment string, duration *models.AeonDuration, unitType string) models.AeonUnit {
	return models.AeonUnit{
		Start:    startDateTime,
		Stop:     stopDateTime,
		Duration: duration,
		Type:     unitType,
		Comment:  comment,
	}
}
