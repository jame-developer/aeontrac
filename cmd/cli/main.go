package main

import (
	"fmt"
	"github.com/jame-developer/aeontrac/configuration"
	"github.com/jame-developer/aeontrac/pkg/commands"
	"github.com/jame-developer/aeontrac/pkg/reporting"
	"github.com/jame-developer/aeontrac/pkg/repositories"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
)

// getXDGPath returns the path for the given environment variable or the fallback value
func getXDGPath(envVar string, fallback string) string {
	value, exists := os.LookupEnv(envVar)
	if !exists {
		value = filepath.Join(os.Getenv("HOME"), fallback)
	}
	return value
}

// getAppFolders returns the configuration and data folders for the application
func getAppFolders() (configFolder, dataFolder string, err error) {
	configPath := getXDGPath("XDG_CONFIG_HOME", ".config")
	dataPath := getXDGPath("XDG_DATA_HOME", ".local/share")

	appName := "aeontrac"
	configFolder = filepath.Join(configPath, appName)
	dataFolder = filepath.Join(dataPath, appName)

	if err = os.MkdirAll(configFolder, 0755); err != nil {
		fmt.Println("Error creating config directory:", err)
		return
	}
	if err = os.MkdirAll(dataFolder, 0755); err != nil {
		fmt.Println("Error creating data directory:", err)
		return
	}

	return
}

// Example function to demonstrate serialization of the Go data structure to JSON
func main() {
	configFolder, dataFolder, err2 := getAppFolders()
	if err2 != nil {
		fmt.Println("Error getting application folders:", err2)
		return
	}

	config, err := configuration.LoadConfig(configFolder)
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}
	valdtr := validator.New()
	data, err := repositories.LoadAeonVault(dataFolder, valdtr)
	if err != nil {
		data, err = repositories.NewAeonVault(time.Now().Year(), config.PublicHolidays)
		if err != nil {
			fmt.Println("Error creating new time tracking data:", err)
			return
		}
	}

	_ = repositories.SaveAeonVault(dataFolder, data)
	var rootCmd = &cobra.Command{
		Use:     "",
		Short:   "TimeLord is a time tracking system",
		Version: "0.1",
		Run: func(cmd *cobra.Command, args []string) {
			reporting.PrintTodayReport(config.WorkingHours, &data)
		},
	}

	var startCmd = &cobra.Command{
		Use:   "start [time] [comment]",
		Short: "Start time tracking for a new unit of work",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			commands.StartCommand(args, &data)
			reporting.PrintTodayReport(config.WorkingHours, &data)
		},
	}
	var stopCmd = &cobra.Command{
		Use:   "stop [time] [comment]",
		Short: "Stop time tracking for a unit of work",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			commands.StopCommand(args, config.WorkingHours, &data)
			reporting.PrintTodayReport(config.WorkingHours, &data)
		},
	}
	var addCmd = &cobra.Command{
		Use:   "add [startTime] [stopTime]",
		Short: "Add a time work unit",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			commands.AddTimeWorkUnitCommand(args, config.WorkingHours, &data)
		},
	}
	var quarterlyReportCmd = &cobra.Command{
		Use:   "qrep",
		Short: "Add a time work unit",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			reporting.PrintQuarterlyReport(config.WorkingHours, &data)
		},
	}
	//var offCmd = &cobra.Command{
	//	Use:   "off [start] [duration] [comment]",
	//	Short: "Add time off",
	//	Args:  cobra.RangeArgs(2, 3),
	//	Run: func(cmd *cobra.Command, args []string) {
	//		data.AddTimeOffCommand(args)
	//	},
	//}
	//
	//var vacCmd = &cobra.Command{
	//	Use:   "vac [start] [duration]",
	//	Short: "Add vacation",
	//	Args:  cobra.ExactArgs(2),
	//	Run: func(cmd *cobra.Command, args []string) {
	//		data.AddVacCommand(args)
	//	},
	//}
	//
	//var reportCmd = &cobra.Command{
	//	Use:   "report [duration]",
	//	Short: "Print a report for the given duration",
	//	Args:  cobra.ExactArgs(1),
	//	Run: func(cmd *cobra.Command, args []string) {
	//		data.ReportCommand(args)
	//	},
	//}

	rootCmd.AddCommand(startCmd, stopCmd, addCmd, quarterlyReportCmd /*, offCmd, vacCmd, reportCmd*/)
	for _, subCmd := range rootCmd.Commands() {
		subCmd.Flags().StringVarP(&data.CommandComment, "comment", "c", "", "Comment for the unit of work, in quotes")
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = repositories.SaveAeonVault(dataFolder, data)
	if err != nil {
		fmt.Println("Error saving time tracking data:", err)
		return
	}
}
