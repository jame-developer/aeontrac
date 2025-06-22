package tracking

import (
	"github.com/google/uuid"
	"github.com/jame-developer/aeontrac/configuration"
	"github.com/jame-developer/aeontrac/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestStartTracking(t *testing.T) {
	now := time.Now()
	nowMinusOneHour := now.Add(-1 * time.Hour)
	nowMinus30min := now.Add(-30 * time.Minute)
	nowPlusOneHour := now.Add(1 * time.Hour)
	tests := []struct {
		name          string
		startDateTime *time.Time
		comment       string
		setupFunc     func(a *models.AeonVault) // Function to setup the environment for the test
		expectedError bool
	}{
		{
			name:          "UnitOfWorkAlreadyRunning",
			startDateTime: &now,
			comment:       "Test Comment",
			setupFunc: func(a *models.AeonVault) {
				// Setup your environment here where a unit of work is already running
				dayKey := time.Now().Format(time.DateOnly)
				unitId := uuid.New()
				a.Days[dayKey] = &models.AeonDay{
					Units: map[uuid.UUID]models.AeonUnit{
						uuid.New(): {
							Start: &nowMinusOneHour,
						},
					},
				}
				a.CurrentRunningUnit = &models.AeonCurrentRunningUnit{
					DayKey: dayKey,
					UnitID: unitId,
				}
			},
			expectedError: true,
		},
		{
			name:          "ProvidedTimeIsWithinPreviouslyCompletedUnitOfWork",
			startDateTime: &nowMinus30min,
			comment:       "Test Comment",
			setupFunc: func(a *models.AeonVault) {
				// Setup your environment here where a unit of work is already completed within the provided time
				dayKey := now.Format(time.DateOnly)
				a.Days[dayKey] = &models.AeonDay{
					Units: map[uuid.UUID]models.AeonUnit{
						uuid.New(): {
							Start: &nowMinusOneHour,
							Stop:  &now,
						},
					},
				}
			},
			expectedError: true,
		},
		{
			name:          "StartedTrackingInTheFuture",
			startDateTime: &nowPlusOneHour,
			comment:       "Test Comment",
			setupFunc:     func(a *models.AeonVault) {},
			expectedError: true,
		},
		{
			name:          "SuccessfullyStartedTrackingForNotExistingDay",
			startDateTime: &now,
			comment:       "Test Comment",
			setupFunc:     func(a *models.AeonVault) {},
			expectedError: false,
		},
		{
			name:          "SuccessfullyStartedTrackingForExistingDay",
			startDateTime: &now,
			comment:       "Test Comment",
			setupFunc: func(a *models.AeonVault) {
				// Setup your environment here where a unit of work is already running
				dayKey := now.Format(time.DateOnly)
				a.Days[dayKey] = &models.AeonDay{
					Units: nil,
				}
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			a := &models.AeonVault{
				Days: make(map[string]*models.AeonDay),
			}
			tt.setupFunc(a)

			// Call startTracking
			err := StartTracking(tt.startDateTime, tt.comment, a)

			// Check the error
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStopTracking(t *testing.T) {
	now := time.Now()
	testDayKey := now.Format(time.DateOnly)
	nowMinusOneHour := now.Add(-1 * time.Hour)
	nowPlusOneHour := now.Add(1 * time.Hour)
	testWorkingHoursConfig := configuration.WorkingHoursConfig{
		Enabled:  true,
		WorkDay:  &models.AeonDuration{Duration: 8 * time.Hour},
		WorkWeek: &models.AeonDuration{Duration: 40 * time.Hour},
	}
	tests := []struct {
		name               string
		stopDateTime       *time.Time
		workingHoursConfig configuration.WorkingHoursConfig
		setupFunc          func(a *models.AeonVault) // Function to setup the environment for the test
		expectedError      bool
		expectedTotal      time.Duration
		expectedOvertime   time.Duration
	}{
		{
			name:               "NoUnitOfWorkIsRunning",
			stopDateTime:       &now,
			workingHoursConfig: testWorkingHoursConfig,
			setupFunc:          func(a *models.AeonVault) {},
			expectedError:      true,
		},
		{
			name:               "ProvidedTimeIsBeforeStartTime",
			stopDateTime:       &nowMinusOneHour,
			workingHoursConfig: testWorkingHoursConfig,
			setupFunc: func(a *models.AeonVault) {
				// Setup your environment here where a unit of work is already running
				dayKey := testDayKey
				unitId := uuid.New()
				a.Days[dayKey] = &models.AeonDay{
					Units: map[uuid.UUID]models.AeonUnit{
						unitId: {
							Start: &now,
						},
					},
				}
				a.CurrentRunningUnit = &models.AeonCurrentRunningUnit{
					DayKey: dayKey,
					UnitID: unitId,
				}
			},
			expectedError: true,
		},
		{
			name:               "SuccessfullyStoppedTrackingWithProvidedTime",
			stopDateTime:       &nowPlusOneHour,
			workingHoursConfig: testWorkingHoursConfig,
			expectedOvertime:   -6 * time.Hour,
			expectedTotal:      2 * time.Hour,
			setupFunc: func(a *models.AeonVault) {
				// Setup your environment here where a unit of work is already running
				dayKey := testDayKey
				unitId := uuid.New()
				a.Days[dayKey] = &models.AeonDay{
					Units: map[uuid.UUID]models.AeonUnit{
						unitId: {
							Start: &nowMinusOneHour,
						},
					},
				}
				a.CurrentRunningUnit = &models.AeonCurrentRunningUnit{
					DayKey: dayKey,
					UnitID: unitId,
				}
			},
			expectedError: false,
		},
		{
			name:               "SuccessfullyStoppedTrackingWithoutProvidedTime",
			stopDateTime:       nil,
			workingHoursConfig: testWorkingHoursConfig,
			expectedOvertime:   -7 * time.Hour,
			expectedTotal:      1 * time.Hour,
			setupFunc: func(a *models.AeonVault) {
				// Setup your environment here where a unit of work is already running
				dayKey := testDayKey
				unitId := uuid.New()
				a.Days[dayKey] = &models.AeonDay{
					Units: map[uuid.UUID]models.AeonUnit{
						unitId: {
							Start: &nowMinusOneHour,
						},
					},
				}
				a.CurrentRunningUnit = &models.AeonCurrentRunningUnit{
					DayKey: dayKey,
					UnitID: unitId,
				}
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			a := &models.AeonVault{
				Days: make(map[string]*models.AeonDay),
			}
			tt.setupFunc(a)

			// Call stopTracking
			err := StopTracking(tt.stopDateTime, tt.workingHoursConfig, a)

			// Check the error
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, a.Days[testDayKey].TotalHours.Duration >= tt.expectedTotal, "Total hours expected: >= %s, got: %s", tt.expectedTotal, a.Days[testDayKey].TotalHours.Duration)
				assert.True(t, a.Days[testDayKey].OvertimeHours.Duration >= tt.expectedOvertime, "Overtime hours expected: >= %s, got: %s", tt.expectedOvertime, a.Days[testDayKey].OvertimeHours.Duration)
			}
		})
	}
}

func TestAddTimeWorkUnit(t *testing.T) {
	testPastStartTime, _ := time.Parse(time.RFC3339, "2020-02-05T08:00:00Z")
	testPastStopTime, _ := time.Parse(time.RFC3339, "2020-02-05T12:00:00Z")
	testStartTime, _ := time.Parse(time.RFC3339, "2020-02-05T13:00:00Z")
	testStopTime, _ := time.Parse(time.RFC3339, "2020-02-05T17:00:00Z")
	testOverlappingStartTime, _ := time.Parse(time.RFC3339, "2020-02-05T11:00:00Z")
	testOverlappingStopTime, _ := time.Parse(time.RFC3339, "2020-02-05T13:30:00Z")
	testDayKey := testStartTime.Format(time.DateOnly)
	testWeekendStartTime, _ := time.Parse(time.RFC3339, "2020-02-08T13:00:00Z")
	testWeekendStopTime, _ := time.Parse(time.RFC3339, "2020-02-08T17:00:00Z")
	testWorkingHoursConfig := configuration.WorkingHoursConfig{
		Enabled:  true,
		WorkDay:  &models.AeonDuration{Duration: 8 * time.Hour},
		WorkWeek: &models.AeonDuration{Duration: 40 * time.Hour},
	}
	tests := []struct {
		name               string
		startDateTime      *time.Time
		stopDateTime       *time.Time
		workingHoursConfig configuration.WorkingHoursConfig
		comment            string
		setupFunc          func(a *models.AeonVault) // Function to setup the environment for the test
		expectedError      bool
		expectedTotal      time.Duration
		expectedOvertime   time.Duration
	}{
		{
			name:               "StopTimeBeforeStartTime",
			startDateTime:      &testStopTime,
			stopDateTime:       &testStartTime,
			workingHoursConfig: testWorkingHoursConfig,
			comment:            "Test Comment",
			setupFunc:          func(a *models.AeonVault) {},
			expectedError:      true,
		},
		{
			name:               "StartTimeWithinCompletedUnit",
			startDateTime:      &testOverlappingStartTime,
			stopDateTime:       &testStopTime,
			workingHoursConfig: testWorkingHoursConfig,
			comment:            "Test Comment",
			setupFunc: func(a *models.AeonVault) {
				// Setup your environment here where a unit of work is already completed within the provided start time
				dayKey := testDayKey
				a.Days[dayKey] = &models.AeonDay{
					Units: map[uuid.UUID]models.AeonUnit{
						uuid.New(): {
							Start: &testPastStartTime,
							Stop:  &testPastStopTime,
						},
					},
				}
			},
			expectedError: true,
		},
		{
			name:               "StopTimeWithinCompletedUnit",
			startDateTime:      &testPastStartTime,
			stopDateTime:       &testOverlappingStopTime,
			workingHoursConfig: testWorkingHoursConfig,
			comment:            "Test Comment",
			setupFunc: func(a *models.AeonVault) {
				// Setup your environment here where a unit of work is already completed within the provided stop time
				dayKey := testDayKey
				a.Days[dayKey] = &models.AeonDay{
					Units: map[uuid.UUID]models.AeonUnit{
						uuid.New(): {
							Start: &testStartTime,
							Stop:  &testStopTime,
						},
					},
				}
			},
			expectedError: true,
		},
		{
			name:               "SuccessfullyAddedTimeWorkUnitDayNotExisting",
			startDateTime:      &testStartTime,
			stopDateTime:       &testStopTime,
			workingHoursConfig: testWorkingHoursConfig,
			comment:            "Test Comment",
			setupFunc:          func(a *models.AeonVault) {},
			expectedError:      false,
			expectedTotal:      4 * time.Hour,
			expectedOvertime:   -4 * time.Hour,
		},
		{
			name:               "SuccessfullyAddedWeekendTimeWorkUnitDayNotExisting",
			startDateTime:      &testWeekendStartTime,
			stopDateTime:       &testWeekendStopTime,
			workingHoursConfig: testWorkingHoursConfig,
			comment:            "Test Comment",
			setupFunc:          func(a *models.AeonVault) {},
			expectedError:      false,
			expectedTotal:      4 * time.Hour,
			expectedOvertime:   4 * time.Hour,
		},
		{
			name:               "SuccessfullyAddedTimeWorkUnitDayExisting",
			startDateTime:      &testStartTime,
			stopDateTime:       &testStopTime,
			workingHoursConfig: testWorkingHoursConfig,
			comment:            "Test Comment",
			setupFunc: func(a *models.AeonVault) {
				// Setup your environment here where a unit of work is already running
				dayKey := testDayKey
				unitId := uuid.New()
				a.Days[dayKey] = &models.AeonDay{
					TotalHours:    &models.AeonDuration{Duration: 4 * time.Hour},
					OvertimeHours: &models.AeonDuration{Duration: -4 * time.Hour},
					Units: map[uuid.UUID]models.AeonUnit{
						unitId: {
							Start: &testPastStartTime,
							Stop:  &testPastStopTime,
							Duration: &models.AeonDuration{ // 4 hours
								Duration: testPastStopTime.Sub(testPastStartTime),
							},
						},
					},
				}
			},
			expectedError:    false,
			expectedTotal:    8 * time.Hour,
			expectedOvertime: 0 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			a := &models.AeonVault{
				Days: make(map[string]*models.AeonDay),
			}
			tt.setupFunc(a)

			// Call addTimeWorkUnit
			err := AddTimeWorkUnit(tt.startDateTime, tt.stopDateTime, tt.comment, tt.workingHoursConfig, a)

			// Check the error
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				dayKeyUnderTest := tt.startDateTime.Format(time.DateOnly)
				assert.True(t, a.Days[dayKeyUnderTest].TotalHours.Duration == tt.expectedTotal, "Total hours expected: %s, got: %s", tt.expectedTotal, a.Days[dayKeyUnderTest].TotalHours.Duration)
				assert.True(t, a.Days[dayKeyUnderTest].OvertimeHours.Duration == tt.expectedOvertime, "Overtime hours expected: %s, got: %s", tt.expectedOvertime, a.Days[dayKeyUnderTest].OvertimeHours.Duration)
			}
		})
	}
}

func TestAddTimeCompensatoryUnit(t *testing.T) {
	testPastStartTime, _ := time.Parse(time.RFC3339, "2020-02-05T08:00:00Z")
	testPastStopTime, _ := time.Parse(time.RFC3339, "2020-02-05T12:00:00Z")
	testStartTime, _ := time.Parse(time.RFC3339, "2020-02-05T13:00:00Z")
	testStopTime, _ := time.Parse(time.RFC3339, "2020-02-05T17:00:00Z")
	testOverlappingStartTime, _ := time.Parse(time.RFC3339, "2020-02-05T11:00:00Z")
	testOverlappingStopTime, _ := time.Parse(time.RFC3339, "2020-02-05T13:30:00Z")
	testDayKey := testStartTime.Format(time.DateOnly)
	testWeekendStartTime, _ := time.Parse(time.RFC3339, "2020-02-08T13:00:00Z")
	testWeekendStopTime, _ := time.Parse(time.RFC3339, "2020-02-08T17:00:00Z")
	testWorkingHoursConfig := configuration.WorkingHoursConfig{
		Enabled:  true,
		WorkDay:  &models.AeonDuration{Duration: 8 * time.Hour},
		WorkWeek: &models.AeonDuration{Duration: 40 * time.Hour},
	}
	tests := []struct {
		name               string
		startDateTime      *time.Time
		stopDateTime       *time.Time
		workingHoursConfig configuration.WorkingHoursConfig
		comment            string
		setupFunc          func(a *models.AeonVault) // Function to setup the environment for the test
		expectedError      bool
		expectedTotal      time.Duration
		expectedOvertime   time.Duration
	}{
		{
			name:               "StopTimeBeforeStartTime",
			startDateTime:      &testStopTime,
			stopDateTime:       &testStartTime,
			workingHoursConfig: testWorkingHoursConfig,
			comment:            "Test Comment",
			setupFunc:          func(a *models.AeonVault) {},
			expectedError:      true,
		},
		{
			name:               "StartTimeWithinCompletedUnit",
			startDateTime:      &testOverlappingStartTime,
			stopDateTime:       &testStopTime,
			workingHoursConfig: testWorkingHoursConfig,
			comment:            "Test Comment",
			setupFunc: func(a *models.AeonVault) {
				// Setup your environment here where a unit of work is already completed within the provided start time
				dayKey := testDayKey
				a.Days[dayKey] = &models.AeonDay{
					Units: map[uuid.UUID]models.AeonUnit{
						uuid.New(): {
							Start: &testPastStartTime,
							Stop:  &testPastStopTime,
						},
					},
				}
			},
			expectedError: true,
		},
		{
			name:               "StopTimeWithinCompletedUnit",
			startDateTime:      &testPastStartTime,
			stopDateTime:       &testOverlappingStopTime,
			workingHoursConfig: testWorkingHoursConfig,
			comment:            "Test Comment",
			setupFunc: func(a *models.AeonVault) {
				// Setup your environment here where a unit of work is already completed within the provided stop time
				dayKey := testDayKey
				a.Days[dayKey] = &models.AeonDay{
					Units: map[uuid.UUID]models.AeonUnit{
						uuid.New(): {
							Start: &testStartTime,
							Stop:  &testStopTime,
						},
					},
				}
			},
			expectedError: true,
		},
		{
			name:               "SuccessfullyAddedTimeWorkUnitDayNotExisting",
			startDateTime:      &testStartTime,
			stopDateTime:       &testStopTime,
			workingHoursConfig: testWorkingHoursConfig,
			comment:            "Test Comment",
			setupFunc:          func(a *models.AeonVault) {},
			expectedError:      false,
			expectedTotal:      -4 * time.Hour,
			expectedOvertime:   -12 * time.Hour,
		},
		{
			name:               "SuccessfullyAddedWeekendTimeWorkUnitDayNotExisting",
			startDateTime:      &testWeekendStartTime,
			stopDateTime:       &testWeekendStopTime,
			workingHoursConfig: testWorkingHoursConfig,
			comment:            "Test Comment",
			setupFunc:          func(a *models.AeonVault) {},
			expectedError:      true,
		},
		{
			name:               "SuccessfullyAddedTimeWorkUnitDayExisting",
			startDateTime:      &testStartTime,
			stopDateTime:       &testStopTime,
			workingHoursConfig: testWorkingHoursConfig,
			comment:            "Test Comment",
			setupFunc: func(a *models.AeonVault) {
				// Setup your environment here where a unit of work is already running
				dayKey := testDayKey
				unitId := uuid.New()
				a.Days[dayKey] = &models.AeonDay{
					TotalHours:    &models.AeonDuration{Duration: 4 * time.Hour},
					OvertimeHours: &models.AeonDuration{Duration: -4 * time.Hour},
					Units: map[uuid.UUID]models.AeonUnit{
						unitId: {
							Start: &testPastStartTime,
							Stop:  &testPastStopTime,
							Duration: &models.AeonDuration{ // 4 hours
								Duration: testPastStopTime.Sub(testPastStartTime),
							},
						},
					},
				}
			},
			expectedError:    false,
			expectedTotal:    0 * time.Hour,
			expectedOvertime: -8 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			a := &models.AeonVault{
				Days: make(map[string]*models.AeonDay),
			}
			tt.setupFunc(a)

			// Call addTimeWorkUnit
			err := addTimeCompensatoryUnit(tt.startDateTime, tt.stopDateTime, tt.comment, tt.workingHoursConfig, a)

			// Check the error
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				dayKeyUnderTest := tt.startDateTime.Format(time.DateOnly)
				assert.True(t, a.Days[dayKeyUnderTest].TotalHours.Duration == tt.expectedTotal, "Total hours expected: %s, got: %s", tt.expectedTotal, a.Days[dayKeyUnderTest].TotalHours.Duration)
				assert.True(t, a.Days[dayKeyUnderTest].OvertimeHours.Duration == tt.expectedOvertime, "Overtime hours expected: %s, got: %s", tt.expectedOvertime, a.Days[dayKeyUnderTest].OvertimeHours.Duration)
			}
		})
	}
}
