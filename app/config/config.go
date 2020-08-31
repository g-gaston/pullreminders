package config

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

var (
	ErrNotExist = errNotExist() // "file does not exist"
)

func errNotExist() error { return os.ErrNotExist }

type Config struct {
	Username    string `yaml:"login"` //These two annotations are needed for backwards compatibility
	AccessToken string `yaml:"token"`
}

func NewFromFile(configFileName, configFolderPath string) (config *Config, err error) {
	configFilePath, err := buildConfigFilePath(configFolderPath, configFileName)
	if err != nil {
		return nil, errors.WithMessagef(err, "error building the path for config file [%s] on path [%s]", configFileName, configFolderPath)
	}

	data, err := readFile(configFilePath)
	if err != nil {
		return nil, errors.WithMessagef(err, "error reading config file [%s] on path [%s]", configFileName, configFolderPath)
	}

	config = &Config{}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, errors.WithMessagef(err, "error reading yaml file [%s]", configFilePath)
	}

	return config, nil
}

func StoreInFile(username, accessToken, configFileName, configFolderPath string) (config *Config, err error) {
	config = &Config{
		Username:    username,
		AccessToken: accessToken,
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return nil, errors.Wrap(err, "error when marshalling config")
	}

	expandedPath, err := expandPath(configFolderPath)
	if err != nil {
		return nil, errors.WithMessagef(err, "error expanding the path for config file [%s] on path [%s]", configFileName, configFolderPath)
	}

	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		err = os.MkdirAll(expandedPath, os.ModePerm)

		if err != nil {
			return nil, errors.WithMessagef(err, "error creating dir [%s]", expandedPath)
		}
	}

	configFilePath, err := absoluteFilePath(expandedPath, configFileName)
	if err != nil {
		return nil, errors.WithMessagef(err, "error building the path for config file [%s] on path [%s]", configFileName, configFolderPath)
	}

	err = ioutil.WriteFile(configFilePath, data, 0644)
	if err != nil {
		return nil, errors.WithMessagef(err, "error writting config file to [%s]", configFilePath)
	}

	return config, nil
}

func buildConfigFilePath(folderPath, fileName string) (string, error) {
	expandedPath, err := expandPath(folderPath)
	if err != nil {
		return "", err
	}

	return absoluteFilePath(expandedPath, fileName)
}

func absoluteFilePath(folderPath, fileName string) (string, error) {
	abs, err := filepath.Abs(path.Join(folderPath, fileName))
	if err != nil {
		return "", err
	}
	return abs, nil
}

func expandPath(folderPath string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", errors.WithMessagef(err, "error getting user current directory while expanding path")
	}
	homeDir := usr.HomeDir

	// I'm surprised this is not part of the standard library, maybe there is a better way of doing it?
	if folderPath == "~" {
		folderPath = homeDir
	} else if strings.HasPrefix(folderPath, "~/") {
		folderPath = filepath.Join(homeDir, folderPath[2:])
	}
	return folderPath, nil
}

func readFile(filePath string) (data []byte, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()

	data, err = ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return data, nil
}
