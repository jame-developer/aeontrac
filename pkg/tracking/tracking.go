package tracking

import (
	"time"

	"github.com/google/uuid"
	"github.com/jame-developer/aeontrac/configuration"
	"github.com/jame-developer/aeontrac/pkg/errors"
	"github.com/jame-developer/aeontrac/pkg/models"
	"github.com/jame-developer/aeontrac/pkg/repositories"
)

// StartTracking starts tracking a new unit of work. If a unit of work is already running, an error is returned.
// An error is returned if the provided time is after the current time or within a previously completed unit of work.
// If the provided time is not provided, the current time is used.
// If the provided comment is not provided, the empty string is used.
// The used type is alway "WORK".
func StartTracking(startDateTime *time.Time, comment string, a *models.AeonVault) error {
	newTrackingStart := time.Now()
	if startDateTime != nil {
		newTrackingStart = *startDateTime
	}
	if newTrackingStart.After(time.Now()) {
		return errors.ErrTimeInFuture
	}
	if a.CurrentRunningUnit != nil {
		return errors.ErrUnitOfWorkRunning
	}
	dayKey := newTrackingStart.Format(time.DateOnly)
	newUnitID := uuid.New()
	newUnit := repositories.NewAeonUnit(&newTrackingStart, nil, comment, nil, repositories.WorkType)
	if day, ok := a.Days[dayKey]; ok {
		if day.Units == nil {
			day.Units = make(map[uuid.UUID]models.AeonUnit)
			a.Days[dayKey] = day
		}
	} else {
		a.Days[dayKey] = repositories.NewAoenDay(newTrackingStart)
	}
	for _, unit := range a.Days[dayKey].Units {
		if unit.Stop != nil && newTrackingStart.After(*unit.Start) && newTrackingStart.Before(*unit.Stop) {
			return errors.ErrTimeWithinCompletedUnit
		}
	}
	a.Days[dayKey].Units[newUnitID] = newUnit
	a.CurrentRunningUnit = &models.AeonCurrentRunningUnit{
		DayKey: dayKey,
		UnitID: newUnitID,
	}

	return nil
}

// StopTracking stops the currently running unit of work.
// If no unit of work is running, an error is returned.
// If the provided time is before the start time of the unit of work, an error is returned.
// If the provided time is not provided, the current time is used.
func StopTracking(stopDateTime *time.Time, workingHoursConfig configuration.WorkingHoursConfig, a *models.AeonVault) error {
	if a.CurrentRunningUnit == nil {
		return errors.ErrNoUnitOfWorkRunning
	}
	stopTime := time.Now()
	if stopDateTime != nil {
		stopTime = *stopDateTime
	}
	if stopTime.Before(*a.Days[a.CurrentRunningUnit.DayKey].Units[a.CurrentRunningUnit.UnitID].Start) {
		return errors.ErrStopTimeBeforeStartTime
	}
	runningUnit := a.Days[a.CurrentRunningUnit.DayKey].Units[a.CurrentRunningUnit.UnitID]
	runningUnit.Stop = &stopTime
	runningUnit.Duration = &models.AeonDuration{
		Duration: runningUnit.Stop.Sub(*runningUnit.Start),
	}
	a.Days[a.CurrentRunningUnit.DayKey].Units[a.CurrentRunningUnit.UnitID] = runningUnit
	currentDay := a.Days[a.CurrentRunningUnit.DayKey]
	if currentDay.TotalHours == nil {
		currentDay.TotalHours = &models.AeonDuration{}
	}
	if currentDay.OvertimeHours == nil {
		currentDay.OvertimeHours = &models.AeonDuration{}
	}
	totalDuration, overtimeDuration := calculateDayWorkDurations(currentDay, &runningUnit, workingHoursConfig)
	currentDay.TotalHours.Duration = totalDuration
	currentDay.OvertimeHours.Duration = overtimeDuration
	a.Days[a.CurrentRunningUnit.DayKey] = currentDay
	a.CurrentRunningUnit = nil
	return nil
}

// AddTimeWorkUnit adds a new unit of work with the provided start and stop times.
// If the provided stop time is before the start time, an error is returned.
// If the provided start time is within a previously completed unit of work, an error is returned.
// If the provided stop time is within a previously completed unit of work, an error is returned.
func AddTimeWorkUnit(startDateTime, stopDateTime *time.Time, comment string, workingHoursConfig configuration.WorkingHoursConfig, a *models.AeonVault) error {
	if startDateTime.After(*stopDateTime) {
		return errors.ErrStopTimeBeforeStartTime
	}
	dayKey := startDateTime.Format(time.DateOnly)
	newUnitID := uuid.New()
	newUnit := repositories.NewAeonUnit(startDateTime, stopDateTime, comment, &models.AeonDuration{Duration: stopDateTime.Sub(*startDateTime)}, repositories.WorkType)
	if day, ok := a.Days[dayKey]; ok {
		if day.Units == nil {
			day.Units = make(map[uuid.UUID]models.AeonUnit)
			a.Days[dayKey] = day
		}
	} else {
		a.Days[dayKey] = repositories.NewAoenDay(*startDateTime)
	}
	for _, unit := range a.Days[dayKey].Units {
		if (unit.Stop != nil && startDateTime.After(*unit.Start) && startDateTime.Before(*unit.Stop)) ||
			(unit.Stop != nil && stopDateTime.After(*unit.Start) && stopDateTime.Before(*unit.Stop)) {
			return errors.ErrTimeWithinCompletedUnit
		}
	}
	currentDay := a.Days[dayKey]
	currentDay.Units[newUnitID] = newUnit
	if currentDay.TotalHours == nil {
		currentDay.TotalHours = &models.AeonDuration{}
	}
	if currentDay.OvertimeHours == nil {
		currentDay.OvertimeHours = &models.AeonDuration{}
	}
	totalDuration, overtimeDuration := calculateDayWorkDurations(currentDay, &newUnit, workingHoursConfig)
	currentDay.TotalHours.Duration = totalDuration
	currentDay.OvertimeHours.Duration = overtimeDuration
	a.Days[dayKey] = currentDay
	return nil
}

// addTimeCompensatoryUnit adds a new unit of compensatory time with the provided start and stop times.
// If the provided stop time is before the start time, an error is returned.
// If the provided start time is within a previously completed unit of work, an error is returned.
// If the provided stop time is within a previously completed unit of work, an error is returned.
func addTimeCompensatoryUnit(startDateTime, stopDateTime *time.Time, comment string, workingHoursConfig configuration.WorkingHoursConfig, a *models.AeonVault) error {
	if startDateTime.After(*stopDateTime) {
		return errors.ErrStopTimeBeforeStartTime
	}

	dayKey := startDateTime.Format(time.DateOnly)
	newUnitID := uuid.New()
	newUnit := repositories.NewAeonUnit(startDateTime, stopDateTime, comment, &models.AeonDuration{Duration: stopDateTime.Sub(*startDateTime)}, repositories.CompensatoryType)
	if day, ok := a.Days[dayKey]; ok {
		if day.Units == nil {
			day.Units = make(map[uuid.UUID]models.AeonUnit)
			a.Days[dayKey] = day
		}
	} else {
		a.Days[dayKey] = repositories.NewAoenDay(*startDateTime)
	}
	currentDay := a.Days[dayKey]
	if currentDay.VacationDay || currentDay.PublicHoliday || currentDay.WeekEnd {
		return errors.ErrCompensationOnNonWorkDay
	}
	for _, unit := range a.Days[dayKey].Units {
		if (unit.Stop != nil && startDateTime.After(*unit.Start) && startDateTime.Before(*unit.Stop)) ||
			(unit.Stop != nil && stopDateTime.After(*unit.Start) && stopDateTime.Before(*unit.Stop)) {
			return errors.ErrTimeWithinCompletedUnit
		}
	}
	currentDay.Units[newUnitID] = newUnit
	if currentDay.TotalHours == nil {
		currentDay.TotalHours = &models.AeonDuration{}
	}
	if currentDay.OvertimeHours == nil {
		currentDay.OvertimeHours = &models.AeonDuration{}
	}
	totalDuration, overtimeDuration := calculateDayCompensatoryDurations(currentDay, &newUnit, workingHoursConfig)
	currentDay.TotalHours.Duration = totalDuration
	currentDay.OvertimeHours.Duration = overtimeDuration
	a.Days[dayKey] = currentDay
	return nil
}

// calculateDayWorkDurations calculates the total and overtime hours for a day.
func calculateDayWorkDurations(currentDay *models.AeonDay, newUnit *models.AeonUnit, workingHoursConfig configuration.WorkingHoursConfig) (time.Duration, time.Duration) {
	totalHours := currentDay.TotalHours.Duration
	overtimeHours := currentDay.OvertimeHours.Duration
	totalHours += newUnit.Duration.Duration
	if workingHoursConfig.Enabled {
		if currentDay.VacationDay || currentDay.PublicHoliday || currentDay.WeekEnd {
			overtimeHours = totalHours
		} else {
			overtimeHours = totalHours - workingHoursConfig.WorkDay.Duration
		}
	}

	return totalHours, overtimeHours
}

// calculateDayCompensatoryDurations calculates the total and overtime hours for a day with compensatory time.
func calculateDayCompensatoryDurations(currentDay *models.AeonDay, newUnit *models.AeonUnit, workingHoursConfig configuration.WorkingHoursConfig) (time.Duration, time.Duration) {
	totalHours := currentDay.TotalHours.Duration
	overtimeHours := currentDay.OvertimeHours.Duration
	totalHours -= newUnit.Duration.Duration
	if workingHoursConfig.Enabled {
		overtimeHours = totalHours - workingHoursConfig.WorkDay.Duration
	}

	return totalHours, overtimeHours
}
