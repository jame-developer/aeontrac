package aeontrac

import (
	"github.com/jame-developer/aeontrac/pkg/models"
	"time"
)

type (
	// PublicHolidaysConfig represents the configuration for public holidays
	PublicHolidaysConfig struct {
		Enabled bool   `json:"enabled"`
		Country string `json:"country" validate:"iso3166_1_alpha2,omitempty"`
		Region  string `json:"region" validate:"omitempty"`
		APIURL  string `json:"api_url" validate:"url,omitempty"`
	}
	WorkingHoursConfig struct {
		// Whether the working hours are enabled
		Enabled bool `json:"enabled"`
		// Start of the work day
		StartTime time.Time `json:"start_time"`
		// End of the work day
		EndTime time.Time `json:"end_time"`
		// Duration of the lunch break
		LunchBreak *models.AeonDuration `json:"lunch_break"`
		// Duration of the work day
		WorkDay *models.AeonDuration `json:"work_day"`
		// Duration of the work week
		WorkWeek *models.AeonDuration `json:"work_week"`
	}
)

func GetDefaultWorkingHoursConfig() WorkingHoursConfig {
	return WorkingHoursConfig{
		Enabled:    true,
		StartTime:  time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:    time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC),
		LunchBreak: &models.AeonDuration{Duration: time.Hour},
		WorkDay:    &models.AeonDuration{Duration: time.Hour * 8},
		WorkWeek:   &models.AeonDuration{Duration: time.Hour * 40},
	}
}
