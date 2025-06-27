package quartly

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// Define your structs for unmarshalling the JSON data
type Day struct {
	IsoWeekNumber     int    `json:"iso_week_number"`
	IsoWeekDay        int    `json:"iso_week_day"`
	PublicHoliday     bool   `json:"public_holiday"`
	PublicHolidayName string `json:"public_holiday_name"`
	VacationDay       bool   `json:"vacation_day"`
	WeekEnd           bool   `json:"week_end"`
	TotalHours        string `json:"total_hours"`
	OvertimeHours     string `json:"overtime_hours"`
}

type AeonDays struct {
	AeonDays map[string]Day `json:"aeon_days"`
}

func Run() {
	// Load JSON file
	file, err := os.Open("aeon_vault.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read JSON file
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal JSON data
	var aeonDays AeonDays
	err = json.Unmarshal(byteValue, &aeonDays)
	if err != nil {
		log.Fatal(err)
	}

	// Calculate the dates for the current month and the two months before
	now := time.Now()
	startOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startOfTwoMonthsBefore := startOfCurrentMonth.AddDate(0, -2, 0)

	// Initialize maps to store total and overtime hours per week
	weekHours := make(map[int]time.Duration)
	weekOvertime := make(map[int]time.Duration)

	// Parse the dates and aggregate hours per week
	for dateStr, day := range aeonDays.AeonDays {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			log.Fatal(err)
		}

		// Consider only the dates in the specified range
		if date.Before(startOfTwoMonthsBefore) || date.After(now) {
			continue
		}

		// Parse total hours and overtime hours
		totalHours, err := time.ParseDuration(day.TotalHours)
		if err != nil {
			totalHours = 0
		}

		overtimeHours, err := time.ParseDuration(day.OvertimeHours)
		if err != nil {
			overtimeHours = 0
		}

		// Aggregate hours by week number
		weekHours[day.IsoWeekNumber] += totalHours
		weekOvertime[day.IsoWeekNumber] += overtimeHours
	}

	// Print total and overtime hours per week
	fmt.Println("Week Number | Total Hours | Overtime Hours")
	fmt.Println("-----------------------------------------")
	for week, total := range weekHours {
		overtime := weekOvertime[week]
		fmt.Printf("Week %d      | %v | %v\n", week, total, overtime)
	}
}