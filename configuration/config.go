package configuration

import (
	"encoding/json"
	"errors"
	"github.com/jame-developer/aeontrac/aeontrac"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	PublicHolidays aeontrac.PublicHolidaysConfig `mapstructure:"public-holidays" json:"public_holidays"`
	WorkingHours   aeontrac.WorkingHoursConfig   `mapstructure:"working-hours" json:"working_hours"`
	Server         ServerConfig                  `mapstructure:"server" json:"server"`
}

// ServerConfig holds the configuration for the HTTP server
type ServerConfig struct {
	Port            int    `mapstructure:"port" json:"port" validate:"required,min=1,max=65535"`
	Host            string `mapstructure:"host" json:"host" validate:"required"`
	ReadTimeout     int    `mapstructure:"read_timeout" json:"read_timeout" validate:"required,min=1"`
	WriteTimeout    int    `mapstructure:"write_timeout" json:"write_timeout" validate:"required,min=1"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout" json:"shutdown_timeout" validate:"required,min=1"`
	APIBasePath     string `mapstructure:"api_base_path" json:"api_base_path" validate:"required"`
	Environment     string `mapstructure:"environment" json:"environment" validate:"required,oneof=development production testing"`
}

func LoadConfig(configPath string) (*Config, error) {
	configFilePath := filepath.Join(configPath, "config.json")

	_, err := os.Stat(configFilePath)
	if errors.Is(err, os.ErrNotExist) {
		defaultConfig := Config{
			PublicHolidays: aeontrac.PublicHolidaysConfig{Enabled: true, Country: "DE"},
			WorkingHours:   aeontrac.GetDefaultWorkingHoursConfig(),
			Server: ServerConfig{
				Port:            8080,
				Host:            "localhost",
				ReadTimeout:     30,
				WriteTimeout:    30,
				ShutdownTimeout: 15,
				APIBasePath:     "/api/v1",
				Environment:     "development",
			},
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
