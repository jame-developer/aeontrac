// Package appcore provides application loading utilities for both CLI and API.
package appcore

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jame-developer/aeontrac/configuration"
	"github.com/jame-developer/aeontrac/pkg/models"
	"github.com/jame-developer/aeontrac/pkg/repositories"
)

var testDir string

// SetTestDir sets a temporary directory for testing purposes.
func SetTestDir(dir string) {
	testDir = dir
}

// LoadApp loads the configuration and AeonVault data, returning config, data, dataFolder, and error.
func LoadApp() (*configuration.Config, *models.AeonVault, string, error) {
	configFolder, dataFolder, err := getAppFolders()
	if err != nil {
		return nil, nil, "", fmt.Errorf("error getting application folders: %w", err)
	}

	config, err := configuration.LoadConfig(configFolder)
	if err != nil {
		return nil, nil, "", fmt.Errorf("error loading configuration: %w", err)
	}

	valdtr := validator.New()
	data, err := repositories.LoadAeonVault(dataFolder, valdtr)
	if err != nil {
		newData, err2 := repositories.NewAeonVault(time.Now().Year(), config.PublicHolidays)
		if err2 != nil {
			return nil, nil, "", fmt.Errorf("error creating new time tracking data: %w", err2)
		}
		data = newData
	}
	// Save to ensure the data file exists
	_ = repositories.SaveAeonVault(dataFolder, data)

	return config, &data, dataFolder, nil
}

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
	if testDir != "" {
		configFolder = filepath.Join(testDir, "config")
		dataFolder = filepath.Join(testDir, "data")
	} else {
		configPath := getXDGPath("XDG_CONFIG_HOME", ".config")
		dataPath := getXDGPath("XDG_DATA_HOME", ".local/share")

		appName := "aeontrac"
		configFolder = filepath.Join(configPath, appName)
		dataFolder = filepath.Join(dataPath, appName)
	}

	if err = os.MkdirAll(configFolder, 0755); err != nil {
		return
	}
	if err = os.MkdirAll(dataFolder, 0755); err != nil {
		return
	}

	return
}

// SaveApp saves the configuration and AeonVault data.
func SaveApp(config *configuration.Config, data *models.AeonVault, dataFolder string) error {
	if err := repositories.SaveAeonVault(dataFolder, *data); err != nil {
		return fmt.Errorf("error saving time tracking data: %w", err)
	}
	return nil
}