package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	projectName string = "pushnotifier"
)

func GetConfigDirPath() (string, error) {
	// Find $XDG_CONFIG_HOME directory.
	UserConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(UserConfigDir, projectName)
	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		return "", nil
	}

	return configDir, nil
}

func GetConfigFilePath() string {
	configDir, err := GetConfigDirPath()
	if err != nil {
		log.Fatal(err)
	}

	fileName := fmt.Sprintf("%s.yaml", projectName)
	configFile := filepath.Join(configDir, fileName)

	return configFile
}

func WriteToConfigFile(content []byte) error {
	configFilePath := GetConfigFilePath()

	err := os.WriteFile(configFilePath, content, 0600)
	if err != nil {
		return err
	}

	return nil
}
