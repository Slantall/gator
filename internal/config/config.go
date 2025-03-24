package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Dburl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {

	dir, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	dat, err := os.ReadFile(dir)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	err = json.Unmarshal(dat, &cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) SetUser(name string) error {
	c.CurrentUserName = name
	dat, err := json.Marshal(c)
	if err != nil {
		return err
	}
	dir, err := getConfigFilePath()
	if err != nil {
		return err
	}
	err = os.WriteFile(dir, dat, 0644)
	if err != nil {
		return err
	}
	return nil
}

func getConfigFilePath() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}
