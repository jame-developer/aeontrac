package aeontrac

import (
	"fmt"
	"os"
	"time"
)

// StopCommand stops time tracking
func (a *AeonVault) StopCommand(args []string, workingHoursConfig WorkingHoursConfig) {
	stopTime, err := parseTimeParam(args, 0)
	if err != nil {
		fmt.Println("Error parsing start time:", err)
		os.Exit(1)
	}
	err = a.stopTracking(&stopTime, workingHoursConfig)
	if err != nil {
		fmt.Printf("Error stopping time tracking: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Time tracking stopped")
}

// StartCommand starts time tracking
func (a *AeonVault) StartCommand(args []string) {
	startTime, err := parseTimeParam(args, 0)
	if err != nil {
		fmt.Println("Error parsing start time:", err)
		os.Exit(1)
	}
	err = a.startTracking(&startTime, "")
	if err != nil {
		fmt.Println("Error starting time tracking:", err)
		os.Exit(1)
	}
}

func (a *AeonVault) AddTimeWorkUnitCommand(args []string, workingHoursConfig WorkingHoursConfig) {
	startTime, err := parseTimeParam(args, 0)
	if err != nil {
		fmt.Println("Error parsing start time:", err)
		os.Exit(1)
	}
	stopTime, err := parseTimeParam(args, 1)
	if err != nil {
		fmt.Println("Error parsing start time:", err)
		os.Exit(1)
	}
	err = a.addTimeWorkUnit(&startTime, &stopTime, "", workingHoursConfig)
	if err != nil {
		fmt.Println("Error adding time off:", err)
		os.Exit(1)
	}
}

// parseTimeParam parses a time parameter from the command line arguments
func parseTimeParam(args []string, expectedPos int) (time.Time, error) {
	paramTime := time.Now()
	if len(args) >= expectedPos+1 {
		parsedTime, err := time.Parse(time.DateOnly+"T"+time.TimeOnly, args[expectedPos])
		if err != nil {
			return time.Time{}, fmt.Errorf("error parsing time ('%s'): %v", args[expectedPos], err)
		}
		paramTime = parsedTime
	}

	return paramTime, nil
}

// parseTimeDuration parses a time duration parameter from the command line arguments
func parseTimeDuration(args []string, expectedPos int) (time.Duration, error) {
	paramTime := time.Hour
	if len(args) >= expectedPos+1 {
		parsedTime, err := time.ParseDuration(args[expectedPos])
		if err != nil {
			return time.Duration(0), fmt.Errorf("error parsing duration ('%s'): %v", args[expectedPos], err)
		}
		paramTime = parsedTime
	}

	return paramTime, nil
}
