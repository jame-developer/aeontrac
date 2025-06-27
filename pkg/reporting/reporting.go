package reporting

import (
	"fmt"
	"github.com/jame-developer/aeontrac/configuration"
	"github.com/jame-developer/aeontrac/pkg/models"
	"log"
	"sort"
	"strings"
	"time"
)

const unitLineTmpl = "%s %s\t%s\t%s"

func PrintTodayReport(workingHoursConfig configuration.WorkingHoursConfig, a *models.AeonVault) {
	today := a.Days[time.Now().Format(time.DateOnly)]
	var reportLines []string
	unitLines := map[int]string{}
	runningDuration := time.Second * 0
	if len(today.Units) > 0 {
		for _, unit := range today.Units {
			if unit.Duration != nil {
				unitLines[int(unit.Start.UnixMicro())] = fmt.Sprintf(unitLineTmpl, " ", unit.Start.Format(time.TimeOnly), unit.Stop.Format(time.TimeOnly), formatDuration(unit.Duration.Duration))
			} else {
				now := time.Now()
				runningDuration = now.Sub(*unit.Start)
				unitLines[int(unit.Start.UnixMicro())] = fmt.Sprintf(unitLineTmpl, "⏱", unit.Start.Format(time.TimeOnly), now.Format(time.TimeOnly), formatDuration(runningDuration))
			}
		}
		keys := make([]int, 0, len(unitLines))
		for k := range unitLines {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		reportLines = append(reportLines, fmt.Sprintf(unitLineTmpl, " ", "Start\t", "End\t", "Duration"))
		for _, key := range keys {
			reportLines = append(reportLines, unitLines[key])
		}
		reportLines = append(reportLines, strings.Repeat("", 32))
	}
	if today.TotalHours == nil {
		fmt.Println("No time tracked today.")
		return
	}
	total := today.TotalHours.Duration + runningDuration
	reportLines = append(reportLines, fmt.Sprintf("TotalHours:\t%s", formatDuration(total)))
	if today.OvertimeHours == nil {
		return
	}
	overtime := total - workingHoursConfig.WorkDay.Duration
	reportLines = append(reportLines, fmt.Sprintf("Overtime:\t%s", formatDuration(overtime)))
	holidayLinesForNextDays := getHolidayLinesForNextDays(7, a)
	if len(holidayLinesForNextDays) > 0 {
		reportLines = append(reportLines, "")
		reportLines = append(reportLines, holidayLinesForNextDays...)
	}
	for _, line := range reportLines {
		println(line)
	}
}

func PrintQuarterlyReport(workingHoursConfig configuration.WorkingHoursConfig, a *models.AeonVault) {
	now := time.Now()
	startOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startOfTwoMonthsBefore := startOfCurrentMonth.AddDate(0, -2, 0)

	// Initialize maps to store total and overtime hours per week
	weekHours := make(map[int]time.Duration)
	weekOvertime := make(map[int]time.Duration)

	// Parse the dates and aggregate hours per week
	for dateStr, day := range a.Days {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			log.Fatal(err)
		}

		// Consider only the dates in the specified range
		if date.Before(startOfTwoMonthsBefore) || date.After(now) {
			continue
		}
		if day.TotalHours == nil {
			continue
		}
		// Aggregate hours by week number
		weekHours[day.IsoWeekNumber] += day.TotalHours.Duration
		weekOvertime[day.IsoWeekNumber] += day.OvertimeHours.Duration
	}

	// Sort week numbers
	weekNumbers := make([]int, 0, len(weekHours))
	for week := range weekHours {
		weekNumbers = append(weekNumbers, week)
	}
	sort.Ints(weekNumbers)
	// Print total and overtime hours per week
	fmt.Println("Week Number | Total Hours  | Overtime Hours")
	fmt.Println("------------------------------------------")
	for _, week := range weekNumbers {
		total := weekHours[week]
		overtime := weekOvertime[week]
		fmt.Printf("Week %-6d | %12s | %12s\n", week, formatDuration(total), formatDuration(overtime))
	}
}

type TodayReportUnit struct {
	Start    string `json:"start"`
	Stop     string `json:"stop"`
	Duration string `json:"duration"`
	Running  bool   `json:"running"`
}

type TodayReport struct {
	Units       []TodayReportUnit `json:"units"`
	TotalHours  string            `json:"total_hours"`
	Overtime    string            `json:"overtime"`
	Holidays    []string          `json:"holidays,omitempty"`
}

func GetTodayReport(workingHoursConfig configuration.WorkingHoursConfig, a *models.AeonVault) TodayReport {
	today := a.Days[time.Now().Format(time.DateOnly)]
	var units []TodayReportUnit
	var runningDuration time.Duration

	for _, unit := range today.Units {
		if unit.Duration != nil {
			units = append(units, TodayReportUnit{
				Start:    unit.Start.Format(time.TimeOnly),
				Stop:     unit.Stop.Format(time.TimeOnly),
				Duration: formatDuration(unit.Duration.Duration),
				Running:  false,
			})
		} else {
			now := time.Now()
			runningDuration = now.Sub(*unit.Start)
			units = append(units, TodayReportUnit{
				Start:    unit.Start.Format(time.TimeOnly),
				Stop:     now.Format(time.TimeOnly),
				Duration: formatDuration(runningDuration),
				Running:  true,
			})
		}
	}

	totalDuration := time.Duration(0)
	if today.TotalHours != nil {
		totalDuration = today.TotalHours.Duration + runningDuration
	}

	overtimeDuration := time.Duration(0)
	if today.OvertimeHours != nil {
		overtimeDuration = totalDuration - workingHoursConfig.WorkDay.Duration
	}

	holidays := getHolidayLinesForNextDays(7, a)

	return TodayReport{
		Units:      units,
		TotalHours: formatDuration(totalDuration),
		Overtime:   formatDuration(overtimeDuration),
		Holidays:   holidays,
	}
}

func getHolidayLinesForNextDays(nextNumberOfDays int, a *models.AeonVault) []string {
	currentDay := time.Now()
	dayInNanoSeconds := int64(time.Hour * 24)
	result := []string{}
	for i := 0; i < nextNumberOfDays; i++ {
		duration := time.Duration(dayInNanoSeconds * int64(i))
		dayKey := currentDay.Add(duration).Format(time.DateOnly)
		if day, ok := a.Days[dayKey]; ok {
			if day.PublicHoliday {
				result = append(result, fmt.Sprintf("%s: %s", dayKey, day.PublicHolidayName))
			}
		}
	}

	return result
}

// formatDuration formats a duration as a string in the format HH:MM:SS
func formatDuration(d time.Duration) string {
	sign := ""
	if d < 0 {
		d = -d
		sign = "-"
	}
	d += 500 * time.Millisecond

	hours := int(d.Hours())
	correctionMultiplier := 1.0
	minutes := int(d.Minutes()*correctionMultiplier) % 60
	seconds := int(d.Seconds()*correctionMultiplier) % 60
	return fmt.Sprintf("%s%02d:%02d:%02d", sign, hours, minutes, seconds)
}
