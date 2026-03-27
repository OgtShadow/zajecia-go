package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"zajecia-go/configloader"
)

type AppConfig struct {
	App struct {
		Name string `yaml:"name"`
		Env  string `yaml:"env"`
		Port int    `yaml:"port"`
	} `yaml:"app"`

	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`

	Features struct {
		MetricsEnabled       bool `yaml:"metrics_enabled"`
		RequestTimeoutSecond int  `yaml:"request_timeout_seconds"`
	} `yaml:"features"`
}

func (c AppConfig) Validate() error {
	if c.App.Name == "" {
		return errors.New("app.name cannot be empty")
	}
	if c.App.Port <= 0 || c.App.Port > 65535 {
		return fmt.Errorf("app.port out of range: %d", c.App.Port)
	}
	if c.Database.Host == "" {
		return errors.New("database.host cannot be empty")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("database.port out of range: %d", c.Database.Port)
	}
	return nil
}

func main() {
	configPath := "config.example.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := configloader.LoadFromYAML[AppConfig](configPath)
	if err != nil {
		if errors.Is(err, configloader.ErrConfigFileNotFound) {
			log.Fatalf("config file not found. provide a valid path as first argument. details: %v", err)
		}
		log.Fatalf("failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config values: %v", err)
	}

	fmt.Printf("Loaded config for app '%s' (%s) on port %d\n", cfg.App.Name, cfg.App.Env, cfg.App.Port)
	fmt.Printf("Database target: %s:%d/%s\n", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	fmt.Printf("Metrics enabled: %t, timeout: %ds\n", cfg.Features.MetricsEnabled, cfg.Features.RequestTimeoutSecond)
}
