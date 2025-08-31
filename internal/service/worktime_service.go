package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jame-developer/aeontrac/pkg/models"
	"github.com/jame-developer/aeontrac/pkg/repositories"
	"github.com/jame-developer/aeontrac/internal/appcore"
)

// AddWorkTimeEntry is a placeholder function for adding a work time entry.
func AddWorkTimeEntry(request models.WorkTimeRequest) (*models.AeonUnit, error) {
	// Load the app core to get config, vault and dataFolder
	config, vault, dataFolder, err := appcore.LoadApp()
	if err != nil {
		return nil, err
	}

	_ = config // config is not used here but loaded for completeness

	// Validate start and stop times
	startTime, err := time.Parse(time.RFC3339, request.Start)
	if err != nil {
		return nil, errors.New("invalid start time format")
	}
	stopTime, err := time.Parse(time.RFC3339, request.Stop)
	if err != nil {
		return nil, errors.New("invalid stop time format")
	}
	if !stopTime.After(startTime) {
		return nil, errors.New("stop time must be after start time")
	}

	// Find or create the AeonDay for the given date
	day, exists := vault.Days[request.Date]
	if !exists {
		day = &models.AeonDay{
			Units: make(map[uuid.UUID]models.AeonUnit),
		}
		vault.Days[request.Date] = day
	}

	// Create a new AeonUnit
	newID := uuid.New()
	duration := stopTime.Sub(startTime)
	newUnit := models.AeonUnit{
		Start:    &startTime,
		Stop:     &stopTime,
		Duration: &models.AeonDuration{Duration: duration},
		Type:     "WORK",
		Comment:  request.Comment,
	}

	// Add the new unit to the day's units map
	day.Units[newID] = newUnit

	// Recalculate TotalHours and OvertimeHours
	var totalDuration time.Duration
	for _, unit := range day.Units {
		if unit.Duration != nil {
			totalDuration += unit.Duration.Duration
		}
	}
	day.TotalHours = &models.AeonDuration{Duration: totalDuration}

	// Calculate overtime as total - 8h if positive
	eightHours := 8 * time.Hour
	overtime := totalDuration - eightHours
	if overtime < 0 {
		overtime = 0
	}
	day.OvertimeHours = &models.AeonDuration{Duration: overtime}

	// Save the vault
	err = repositories.SaveAeonVault(dataFolder, *vault)
	if err != nil {
		return nil, err
	}

	// Return the new unit
	return &newUnit, nil
}