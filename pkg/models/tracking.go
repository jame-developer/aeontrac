package models

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
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
