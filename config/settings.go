package config

import (
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

const settingsPath = "/app/config"

var settingsFiles = map[string]string{
	"development": path.Join(settingsPath, "appConfig.dev.yml"),
	"production":  path.Join(settingsPath, "appConfig.yml"),
}

// Service contains the settings to connect to services
type Service struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

// Settings struct to unmarshal config yml settings
type Settings struct {
	User Service
}

var environment string = os.Getenv("Environment")

// LoadSettings loads the settings from the yml file
func LoadSettings() (*Settings, error) {
	config, err := ioutil.ReadFile(settingsFiles[environment])
	if err != nil {
		return nil, err
	}

	settings := &Settings{}
	err = yaml.Unmarshal(config, settings)
	if err != nil {
		return nil, err
	}

	return settings, nil
}
