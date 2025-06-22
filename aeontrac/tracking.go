package aeontrac

import (
	"github.com/google/uuid"
	"time"
)

// startTracking starts tracking a new unit of work. If a unit of work is already running, an error is returned.
// An error is returned if the provided time is after the current time or within a previously completed unit of work.
// If the provided time is not provided, the current time is used.
// If the provided comment is not provided, the empty string is used.
// The used type is alway "WORK".
func (a *AeonVault) startTracking(startDateTime *time.Time, comment string) error {
	newTrackingStart := time.Now()
	if startDateTime != nil {
		newTrackingStart = *startDateTime
	}
	if newTrackingStart.After(time.Now()) {
		return ErrTimeInFuture
	}
	if a.CurrentRunningUnit != nil {
		return ErrUnitOfWorkRunning
	}
	dayKey := newTrackingStart.Format(time.DateOnly)
	newUnitID := uuid.New()
	newUnit := newAeonUnit(&newTrackingStart, nil, comment, nil, workType)
	if day, ok := a.Days[dayKey]; ok {
		if day.Units == nil {
			day.Units = make(map[uuid.UUID]AeonUnit)
			a.Days[dayKey] = day
		}
	} else {
		a.Days[dayKey] = newAoenDay(*startDateTime)
	}
	for _, unit := range a.Days[dayKey].Units {
		if unit.Stop != nil && newTrackingStart.After(*unit.Start) && newTrackingStart.Before(*unit.Stop) {
			return ErrTimeWithinCompletedUnit
		}
	}
	a.Days[dayKey].Units[newUnitID] = newUnit
	a.CurrentRunningUnit = &AeonCurrentRunningUnit{
		DayKey: dayKey,
		UnitID: newUnitID,
	}

	return nil
}

// stopTracking stops the currently running unit of work.
// If no unit of work is running, an error is returned.
// If the provided time is before the start time of the unit of work, an error is returned.
// If the provided time is not provided, the current time is used.
func (a *AeonVault) stopTracking(stopDateTime *time.Time, workingHoursConfig WorkingHoursConfig) error {
	if a.CurrentRunningUnit == nil {
		return ErrNoUnitOfWorkRunning
	}
	stopTime := time.Now()
	if stopDateTime != nil {
		stopTime = *stopDateTime
	}
	if stopTime.Before(*a.Days[a.CurrentRunningUnit.DayKey].Units[a.CurrentRunningUnit.UnitID].Start) {
		return ErrStopTimeBeforeStartTime
	}
	runningUnit := a.Days[a.CurrentRunningUnit.DayKey].Units[a.CurrentRunningUnit.UnitID]
	runningUnit.Stop = &stopTime
	runningUnit.Duration = &AeonDuration{
		Duration: runningUnit.Stop.Sub(*runningUnit.Start),
	}
	a.Days[a.CurrentRunningUnit.DayKey].Units[a.CurrentRunningUnit.UnitID] = runningUnit
	currentDay := a.Days[a.CurrentRunningUnit.DayKey]
	if currentDay.TotalHours == nil {
		currentDay.TotalHours = &AeonDuration{}
	}
	if currentDay.OvertimeHours == nil {
		currentDay.OvertimeHours = &AeonDuration{}
	}
	totalDuration, overtimeDuration := calculateDayWorkDurations(currentDay, &runningUnit, workingHoursConfig)
	currentDay.TotalHours.Duration = totalDuration
	currentDay.OvertimeHours.Duration = overtimeDuration
	a.Days[a.CurrentRunningUnit.DayKey] = currentDay
	a.CurrentRunningUnit = nil
	return nil
}

// addTimeWorkUnit adds a new unit of work with the provided start and stop times.
// If the provided stop time is before the start time, an error is returned.
// If the provided start time is within a previously completed unit of work, an error is returned.
// If the provided stop time is within a previously completed unit of work, an error is returned.
func (a *AeonVault) addTimeWorkUnit(startDateTime, stopDateTime *time.Time, comment string, workingHoursConfig WorkingHoursConfig) error {
	if startDateTime.After(*stopDateTime) {
		return ErrStopTimeBeforeStartTime
	}
	dayKey := startDateTime.Format(time.DateOnly)
	newUnitID := uuid.New()
	newUnit := newAeonUnit(startDateTime, stopDateTime, comment, &AeonDuration{Duration: stopDateTime.Sub(*startDateTime)}, workType)
	if day, ok := a.Days[dayKey]; ok {
		if day.Units == nil {
			day.Units = make(map[uuid.UUID]AeonUnit)
			a.Days[dayKey] = day
		}
	} else {
		a.Days[dayKey] = newAoenDay(*startDateTime)
	}
	for _, unit := range a.Days[dayKey].Units {
		if (unit.Stop != nil && startDateTime.After(*unit.Start) && startDateTime.Before(*unit.Stop)) ||
			(unit.Stop != nil && stopDateTime.After(*unit.Start) && stopDateTime.Before(*unit.Stop)) {
			return ErrTimeWithinCompletedUnit
		}
	}
	currentDay := a.Days[dayKey]
	currentDay.Units[newUnitID] = newUnit
	if currentDay.TotalHours == nil {
		currentDay.TotalHours = &AeonDuration{}
	}
	if currentDay.OvertimeHours == nil {
		currentDay.OvertimeHours = &AeonDuration{}
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
func (a *AeonVault) addTimeCompensatoryUnit(startDateTime, stopDateTime *time.Time, comment string, workingHoursConfig WorkingHoursConfig) error {
	if startDateTime.After(*stopDateTime) {
		return ErrStopTimeBeforeStartTime
	}

	dayKey := startDateTime.Format(time.DateOnly)
	newUnitID := uuid.New()
	newUnit := newAeonUnit(startDateTime, stopDateTime, comment, &AeonDuration{Duration: stopDateTime.Sub(*startDateTime)}, compensatoryType)
	if day, ok := a.Days[dayKey]; ok {
		if day.Units == nil {
			day.Units = make(map[uuid.UUID]AeonUnit)
			a.Days[dayKey] = day
		}
	} else {
		a.Days[dayKey] = newAoenDay(*startDateTime)
	}
	currentDay := a.Days[dayKey]
	if currentDay.VacationDay || currentDay.PublicHoliday || currentDay.WeekEnd {
		return ErrCompensationOnNonWorkDay
	}
	for _, unit := range a.Days[dayKey].Units {
		if (unit.Stop != nil && startDateTime.After(*unit.Start) && startDateTime.Before(*unit.Stop)) ||
			(unit.Stop != nil && stopDateTime.After(*unit.Start) && stopDateTime.Before(*unit.Stop)) {
			return ErrTimeWithinCompletedUnit
		}
	}
	currentDay.Units[newUnitID] = newUnit
	if currentDay.TotalHours == nil {
		currentDay.TotalHours = &AeonDuration{}
	}
	if currentDay.OvertimeHours == nil {
		currentDay.OvertimeHours = &AeonDuration{}
	}
	totalDuration, overtimeDuration := calculateDayCompensatoryDurations(currentDay, &newUnit, workingHoursConfig)
	currentDay.TotalHours.Duration = totalDuration
	currentDay.OvertimeHours.Duration = overtimeDuration
	a.Days[dayKey] = currentDay
	return nil
}

// calculateDayWorkDurations calculates the total and overtime hours for a day.
func calculateDayWorkDurations(currentDay *AeonDay, newUnit *AeonUnit, workingHoursConfig WorkingHoursConfig) (time.Duration, time.Duration) {
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
func calculateDayCompensatoryDurations(currentDay *AeonDay, newUnit *AeonUnit, workingHoursConfig WorkingHoursConfig) (time.Duration, time.Duration) {
	totalHours := currentDay.TotalHours.Duration
	overtimeHours := currentDay.OvertimeHours.Duration
	totalHours -= newUnit.Duration.Duration
	if workingHoursConfig.Enabled {
		overtimeHours = totalHours - workingHoursConfig.WorkDay.Duration
	}

	return totalHours, overtimeHours
}
