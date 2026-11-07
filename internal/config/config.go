package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (*Config, error) {

	configPath, err := getConfigFilePath()
	if err != nil {
		return &Config{}, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return &Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{}, err
	}
	return &cfg, nil

}

func getConfigFilePath() (string, error) {
	const configFileName = ".gatorconfig.json"
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home + "/" + configFileName, nil
}

func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username
	return write(*c) // or similar
}

func (c *Config) SetDbUrl(dbUrl string) error {
	c.DbUrl = dbUrl
	return write(*c) // or similar
}

func (c *Config) GetDbUrl() string {
	return c.DbUrl
}
func write(cfg Config) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0600)
}

func (c *Config) DeleteUsers() error {
	return write(*c) // or similar
}
