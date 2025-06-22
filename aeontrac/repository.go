package aeontrac

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	dataFileName       = "aeon_vault.json"
	backUpFileNameTmpl = "%d.bak"
	workType           = "WORK"
	compensatoryType   = "COMPENSATORY"
)

type (
	// AeonDay represents all tracking entries for a single day, including the total and overtime hours
	// it also includes information about weekends, public holidays, and vacation days
	// and the week number and week day according to the ISO 8601 standard.
	AeonDay struct {
		IsoWeekNumber     int                    `json:"iso_week_number" validate:"min=1,max=53"`
		IsoWeekDay        int                    `json:"iso_week_day" validate:"min=1,max=7"`
		PublicHoliday     bool                   `json:"public_holiday"`
		PublicHolidayName string                 `json:"public_holiday_name,omitempty" validate:"required_if=PublicHoliday true"`
		VacationDay       bool                   `json:"vacation_day"`
		TotalHours        *AeonDuration          `json:"total_hours,omitempty"`
		OvertimeHours     *AeonDuration          `json:"overtime_hours,omitempty"`
		Units             map[uuid.UUID]AeonUnit `json:"units,omitempty"`
		WeekEnd           bool                   `json:"week_end"`
	}
	// AeonUnit represents a single time tracking entry
	AeonUnit struct {
		Start    *time.Time    `json:"start,omitempty"`
		Stop     *time.Time    `json:"stop,omitempty"`
		Duration *AeonDuration `json:"duration,omitempty"`
		Type     string        `json:"type" validate:"oneof=WORK COMPENSATORY"`
		Comment  string        `json:"comment,omitempty"`
	}
	// AeonVault represents all tracking data
	AeonVault struct {
		Days               map[string]*AeonDay     `json:"aeon_days" validate:"required"`
		CurrentRunningUnit *AeonCurrentRunningUnit `json:"current_running_unit,omitempty"`
		CommandComment     string                  `json:"-"` // CommandComment is used to store the comment for the current command
	}
	AeonCurrentRunningUnit struct {
		DayKey string
		UnitID uuid.UUID
	}
	ReportItem struct {
		TotalHours time.Duration
		MustHours  time.Duration
	}
	AeonDuration struct {
		time.Duration
	}
)

// MarshalJSON converts the duration to a string format for JSON.
func (d *AeonDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON converts a string back to a duration.
func (d *AeonDuration) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	duration, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	d.Duration = duration
	return nil
}

func (r *ReportItem) setMustHours(mustHours time.Duration) {
	r.MustHours = mustHours
}
func (r *ReportItem) addTotalHours(totalHours time.Duration) {
	r.TotalHours += totalHours
}

// LoadAeonVault loads the time tracking data from the provided folder, using the provided public holdidays configuration, if the data does not exist, it creates a new one.
func LoadAeonVault(folder string, validator *validator.Validate) (AeonVault, error) {
	fileName := filepath.Join(folder, dataFileName)
	// Check if the file exists
	if _, err := os.Stat(fileName); err != nil {
		return AeonVault{}, err
	}

	file, err := os.Open(fileName)
	if err != nil {
		return AeonVault{}, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var currentYear AeonVault
	err = json.NewDecoder(file).Decode(&currentYear)
	if err != nil {
		return AeonVault{}, err
	}
	validErr := validator.Struct(&currentYear)
	if validErr != nil {
		return AeonVault{}, validErr
	}
	for _, aeonDay := range currentYear.Days {
		dayValidErr := validator.Struct(aeonDay)
		if dayValidErr != nil {
			return AeonVault{}, dayValidErr
		}
		for _, aeonUnit := range aeonDay.Units {
			unitValidErr := validator.Struct(aeonUnit)
			if unitValidErr != nil {
				return AeonVault{}, unitValidErr
			}
		}

	}
	return currentYear, nil

}

// SaveAeonVault saves the time tracking data to the provided folder
func SaveAeonVault(folder string, data AeonVault) error {

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
func NewAeonVault(year int, publicHolidaysConfig PublicHolidaysConfig) (AeonVault, error) {
	holidays, err := loadHolidays(publicHolidaysConfig, year)
	if err != nil {
		return AeonVault{}, err
	}

	days := map[string]*AeonDay{}
	firstDayOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	for d := firstDayOfYear; d.Year() == year; d = d.AddDate(0, 0, 1) {
		dayKey := d.Format(time.DateOnly)
		newDay := newAoenDay(d)
		if holiday, ok := holidays[dayKey]; ok {
			newDay.PublicHoliday = true
			newDay.PublicHolidayName = holiday.Name
		}

		days[dayKey] = newDay
	}

	return AeonVault{
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

// newAoenDay creates a new AeonDay instance with the provided date
func newAoenDay(date time.Time) *AeonDay {
	_, isoWeekNumber := date.ISOWeek()
	isoDayOfWeekNumber, _ := getIsoDayOfWeek(date)
	isWeekEnd := isoDayOfWeekNumber == 6 || isoDayOfWeekNumber == 7
	return &AeonDay{
		IsoWeekNumber: isoWeekNumber,
		IsoWeekDay:    isoDayOfWeekNumber,
		WeekEnd:       isWeekEnd,
		Units:         map[uuid.UUID]AeonUnit{},
	}
}

// newAeonUnit creates a new AeonUnit instance with the provided parameters
func newAeonUnit(startDateTime, stopDateTime *time.Time, comment string, duration *AeonDuration, unitType string) AeonUnit {
	return AeonUnit{
		Start:    startDateTime,
		Stop:     stopDateTime,
		Duration: duration,
		Type:     unitType,
		Comment:  comment,
	}
}
