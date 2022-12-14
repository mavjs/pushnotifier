/*
Copyright © 2022 Maverick Kaung <mavjs01@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
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
