package models

type WorkTimeRequest struct {
	Date    string `json:"date"`
	Start   string `json:"start"`
	Stop    string `json:"stop"`
	Comment string `json:"comment"`
}

type ReportResponse struct {
	TotalHours    string      `json:"total_hours"`
	OvertimeHours string      `json:"overtime_hours"`
	Days          []DayReport `json:"days"`
}

type DayReport struct {
	Date          string `json:"date"`
	TotalHours    string `json:"total_hours"`
	OvertimeHours string `json:"overtime_hours"`
}
