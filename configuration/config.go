package configuration

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	PublicHolidays PublicHolidaysConfig `mapstructure:"public-holidays" json:"public_holidays"`
	WorkingHours   WorkingHoursConfig   `mapstructure:"working-hours" json:"working_hours"`
}

func LoadConfig(configPath string) (*Config, error) {
	configFilePath := filepath.Join(configPath, "config.json")

	_, err := os.Stat(configFilePath)
	if errors.Is(err, os.ErrNotExist) {
		defaultConfig := Config{
			PublicHolidays: PublicHolidaysConfig{Enabled: true, Country: "DE", APIURL: "https://openholidaysapi.org/PublicHolidays"},
			WorkingHours:   GetDefaultWorkingHoursConfig(),
		}
		bytes, err := json.Marshal(&defaultConfig)
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(configFilePath, bytes, 0644)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	bytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
