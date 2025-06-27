package cli

import (
	"fmt"

	"github.com/jame-developer/aeontrac/pkg/commands"
	"github.com/jame-developer/aeontrac/pkg/reporting"
	"github.com/jame-developer/aeontrac/pkg/repositories"
	"github.com/jame-developer/aeontrac/internal/appcore"
	"github.com/spf13/cobra"
)

// Run initializes and executes the CLI commands.
func Run() error {
	config, data, dataFolder, err := appcore.LoadApp()
	if err != nil {
		return fmt.Errorf("error loading app: %w", err)
	}

	var rootCmd = &cobra.Command{
		Use:     "",
		Short:   "TimeLord is a time tracking system",
		Version: "0.1",
		Run: func(cmd *cobra.Command, args []string) {
			reporting.PrintTodayReport(config.WorkingHours, data)
		},
	}

	var startCmd = &cobra.Command{
		Use:   "start [time] [comment]",
		Short: "Start time tracking for a new unit of work",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			commands.StartCommand(args, data)
			reporting.PrintTodayReport(config.WorkingHours, data)
		},
	}

	var stopCmd = &cobra.Command{
		Use:   "stop [time] [comment]",
		Short: "Stop time tracking for a unit of work",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			commands.StopCommand(args, config.WorkingHours, data)
			reporting.PrintTodayReport(config.WorkingHours, data)
		},
	}

	var addCmd = &cobra.Command{
		Use:   "add [startTime] [stopTime]",
		Short: "Add a time work unit",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			commands.AddTimeWorkUnitCommand(args, config.WorkingHours, data)
		},
	}

	var quarterlyReportCmd = &cobra.Command{
		Use:   "qrep",
		Short: "Add a time work unit",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			reporting.PrintQuarterlyReport(config.WorkingHours, data)
		},
	}

	rootCmd.AddCommand(startCmd, stopCmd, addCmd, quarterlyReportCmd /*, offCmd, vacCmd, reportCmd*/)
	for _, subCmd := range rootCmd.Commands() {
		subCmd.Flags().StringVarP(&data.CommandComment, "comment", "c", "", "Comment for the unit of work, in quotes")
	}

	if err := rootCmd.Execute(); err != nil {
		return err
	}

	err = repositories.SaveAeonVault(dataFolder, *data)
	if err != nil {
		return fmt.Errorf("error saving time tracking data: %w", err)
	}

	return nil
}