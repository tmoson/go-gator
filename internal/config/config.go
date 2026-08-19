package config

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigFile() (*os.File, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(homeDir, configFileName)
	_, err = os.Stat(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			configFile, err := os.Create(configPath)
			if err != nil {
				return nil, err
			}
			return configFile, err
		}
		return nil, err
	}
	// open config file. setting permissions to be read, write, or create, and mode to append so that the
	// file may be edited when opened
	configFile, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE, os.ModeAppend)
	if err != nil {
		return nil, err
	}
	return configFile, nil
}

func (c *Config) write() error {
	configFile, err := getConfigFile()
	if err != nil {
		return err
	}
	defer configFile.Close()
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(c)
	return err
}

func (c *Config) SetUser(currentUser string) error {
	c.CurrentUserName = currentUser
	err := c.write()
	return err
}

func Read() Config {
	var loadedConfig Config
	configFile, err := getConfigFile()
	if err != nil {
		log.Fatal("unable to access config file")
		return Config{}
	}
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&loadedConfig)
	if loadedConfig.DbUrl == "" {
		log.Fatal("ERROR READING FILE...")
	}
	return loadedConfig
}
