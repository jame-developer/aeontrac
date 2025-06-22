package repositories

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jame-developer/aeontrac/aeontrac"
	"github.com/jame-developer/aeontrac/pkg/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNewAeonVault(t *testing.T) {
	// Define your test holidays
	holidaysJSON := `[{"id":"1","startDate":"2024-01-01","endDate":"2024-01-01","type":"Public","name":[{"language":"en","text":"New Year's Day"}],"nationwide":true,"subdivisions":[{"code":"DE","shortName":"Germany"}]},{"id":"2","startDate":"2024-12-25","endDate":"2024-12-25","type":"Public","name":[{"language":"en","text":"Christmas Day"}],"nationwide":true,"subdivisions":[{"code":"DE","shortName":"Germany"}]},{"id":"3","startDate":"2024-12-26","endDate":"2024-12-26","type":"Public","name":[{"language":"en","text":"Second Day of Christmas"}],"nationwide":true,"subdivisions":[{"code":"DE","shortName":"Germany"}]}]`

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := rw.Write([]byte(holidaysJSON))
		if err != nil {
			assert.NoError(t, err)
			return
		}
	}))
	defer server.Close()
	tests := []struct {
		name                 string
		year                 int
		publicHolidaysConfig aeontrac.PublicHolidaysConfig
		workingHoursConfig   aeontrac.WorkingHoursConfig
		expectedError        error
		expectedHolidays     []time.Time
	}{
		{
			name: "ValidInputWithHolidays",
			year: 2024,
			publicHolidaysConfig: aeontrac.PublicHolidaysConfig{
				Enabled: true,
				Country: "DE",
				Region:  "",
				APIURL:  server.URL,
			},
			expectedError: nil,
			expectedHolidays: []time.Time{
				time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 12, 26, 0, 0, 0, 0, time.UTC),
			},
		},

		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vault, err := NewAeonVault(tt.year, tt.publicHolidaysConfig)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vault)
				assert.NotNil(t, vault.Days)
				assert.True(t, len(vault.Days) == 365 || len(vault.Days) == 366)
				for key, value := range vault.Days {
					assert.True(t, value.IsoWeekNumber > 0, "IsoWeekNumber should be greater than 0, for day: %s", key)
					assert.True(t, value.IsoWeekNumber < 54, "IsoWeekNumber should be less than 54, for day: %s", key)
					assert.True(t, value.WeekEnd == (value.IsoWeekDay == 6 || value.IsoWeekDay == 7), "WeekEnd should be true if IsoWeekDay is 6 or 7, for day: %s", key)
					assert.True(t, value.IsoWeekDay > 0, "IsoWeekDay should be greater than 0, for day: %s", key)
					assert.True(t, value.IsoWeekDay < 8, "IsoWeekDay should be less than 8, for day: %s", key)
				}
				for _, holiday := range tt.expectedHolidays {
					dayKey := holiday.Format(time.DateOnly)
					day, exists := vault.Days[dayKey]
					assert.True(t, exists)
					assert.True(t, day.PublicHoliday)
				}
			}
		})
	}
}

func TestLoadAeonVault(t *testing.T) {
	testValidator := validator.New()
	// Define your test holidays
	holidaysJSON := `[{"id":"1","startDate":"2024-01-01","endDate":"2024-01-01","type":"Public","name":[{"language":"en","text":"New Year's Day"}],"nationwide":true,"subdivisions":[{"code":"DE","shortName":"Germany"}]},{"id":"2","startDate":"2024-12-25","endDate":"2024-12-25","type":"Public","name":[{"language":"en","text":"Christmas Day"}],"nationwide":true,"subdivisions":[{"code":"DE","shortName":"Germany"}]},{"id":"3","startDate":"2024-12-26","endDate":"2024-12-26","type":"Public","name":[{"language":"en","text":"Second Day of Christmas"}],"nationwide":true,"subdivisions":[{"code":"DE","shortName":"Germany"}]}]`
	testFolder := "./test_data"
	if _, err := os.Stat(testFolder); err != nil {
		if mkDirErr := os.Mkdir(testFolder, 0755); mkDirErr != nil {
			assert.NoError(t, mkDirErr)
			return
		}
	}
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := rw.Write([]byte(holidaysJSON))
		if err != nil {
			assert.NoError(t, err)
			return
		}
	}))
	defer server.Close()
	tests := []struct {
		name          string
		folder        string
		setupFunc     func(folder string)
		expectedError bool
	}{
		{
			name:   "SuccessfullyLoaded",
			folder: testFolder,
			setupFunc: func(folder string) {
				// Write data to the file
				err := os.WriteFile(folder+"/aeon_vault.json", []byte(`{"aeon_days": {"2024-01-01": {"iso_week_number": 1, "iso_week_day": 1, "public_holiday": true, "public_holiday_name": "New Year's Day", "vacation_day": false, "week_end": false, "units": {}}}}`), 0644)
				if err != nil {
					assert.NoError(t, err)
					return
				}
			},
			expectedError: false,
		},
		{
			name:   "NoDataToLoadNewDataAreCreatedAndSaved",
			folder: testFolder,
			setupFunc: func(folder string) {
				// Remove the file if it exists
				err := os.Remove(folder + "/aeon_vault.json")
				if err != nil {
					assert.NoError(t, err)
					return
				}
			},
			expectedError: true,
		},
		{
			name:   "VaultValidationFails",
			folder: testFolder,
			setupFunc: func(folder string) {
				// Write invalid data to the file
				err := os.WriteFile(folder+"/aeon_vault.json", []byte(`{"aeon_days": null}`), 0644)
				if err != nil {
					assert.NoError(t, err)
					return
				}
			},
			expectedError: true,
		},
		{
			name:   "DaysValidationFails",
			folder: testFolder,
			setupFunc: func(folder string) {
				// Write invalid data to the file
				err := os.WriteFile(folder+"/aeon_vault.json", []byte(`{"aeon_days": {"2024-01-01": {"iso_week_number": 0, "iso_week_day": 100, "public_holiday": true, "public_holiday_name": "New Year's Day", "vacation_day": false, "week_end": false, "units": {}}}}`), 0644)
				if err != nil {
					assert.NoError(t, err)
					return
				}
			},
			expectedError: true,
		},
		{
			name:   "DataCouldNotBeDecoded",
			folder: testFolder,
			setupFunc: func(folder string) {
				// Write non-JSON data to the file
				err := os.WriteFile(folder+"/aeon_vault.json", []byte("not json data"), 0644)
				if err != nil {
					assert.NoError(t, err)
					return
				}
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupFunc(tt.folder)

			// Call LoadAeonVault
			_, err := LoadAeonVault(tt.folder, testValidator)

			// Check the error
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
	_ = os.Remove(testFolder + "/aeon_vault.json")
}

func TestSaveAeonVault(t *testing.T) {
	tests := []struct {
		name          string
		folder        string
		data          models.AeonVault
		expectedError bool
	}{
		{
			name:   "FolderDoesNotExist",
			folder: "./non_existent_folder",
			data: models.AeonVault{
				Days: map[string]*models.AeonDay{},
			},
			expectedError: true,
		},
		{
			name:   "InsufficientPermissions",
			folder: "/root",
			data: models.AeonVault{
				Days: map[string]*models.AeonDay{},
			},
			expectedError: true,
		},
		{
			name:   "SuccessfullySaved",
			folder: "./test_data",
			data: models.AeonVault{
				Days: map[string]*models.AeonDay{
					"2024-01-01": {
						IsoWeekNumber: 1,
						IsoWeekDay:    1,
						PublicHoliday: true,
						Units:         map[uuid.UUID]models.AeonUnit{},
					},
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SaveAeonVault(tt.folder, tt.data)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
