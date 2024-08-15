package app

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type (
	//==============================================================================
	// Structure DBConfig Declaration

	/// Structure for the Database Configuration
	DBConfig struct {
		Host     string `yaml:"host"`
		Name     string `yaml:"name"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	}

	//==============================================================================
	// Structure AppConfig Declaration

	/// Structure for the Application Configuration
	AppConfig struct {
		Component     string   `yaml:"component"`
		Project       string   `yaml:"project"`
		WebRoot       string   `yaml:"web_root"`
		MainDirectory string   `yaml:"main_directory"`
		ConfigFile    string   `yaml:"config_file"`
		DB            DBConfig `yaml:"database"`
	}
)

const CONFIG_FILE string = ".env"

func existsFile(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func findConfigFile(directory string) (string, error) {
	var configFile string
	var err error
	var exists bool = false

	ginMode := gin.Mode()

	//exeDirSep := filepath.FromSlash(exePath)
	parsedDir := strings.ReplaceAll(directory, string(os.PathSeparator), string(os.PathListSeparator))

	fmt.Printf("App - findConfigFile():pth sep: '%c'; lst sep '%c'; prsd dir: %s\n", os.PathListSeparator, os.PathSeparator, parsedDir)

	dirList := filepath.SplitList(parsedDir)

	fmt.Printf("App - findConfigFile(): dir lst (count: %d): %#v\n", len(dirList), dirList)

	for last := len(dirList); !exists && last > 0; last-- {
		dir := strings.Join(dirList[0:last], string(os.PathSeparator))

		if dir == "" {
			dir = string(os.PathSeparator)
		}

		fmt.Printf("App - findConfigFile(): dir (last: %d): '%s'\n", last, dir)

		configFile = path.Join(dir, CONFIG_FILE+"."+ginMode)

		exists = existsFile(configFile)

		fmt.Printf("App - dir: '%s': config 0 '%s'; ex: '%v'\n", dir, configFile, exists)

		if !exists {
			configFile = path.Join(dir, CONFIG_FILE)
			exists = existsFile(configFile)

			fmt.Printf("App - dir: '%s': config 1 '%s'; ex: '%v'\n", dir, configFile, exists)
		}
	}

	if !exists {
		return "", os.ErrNotExist
	}

	return configFile, err
}

func ReadConfigFile() (AppConfig, error) {
	var config AppConfig
	var homeDir string
	var currentDir string
	var exePath string
	var configFile string
	var err error

	homeDir = os.Getenv("HOME")
	currentDir, err = os.Getwd()

	exePath, err = os.Executable()

	if err != nil {
		return config, err
	}

	exePath, err = filepath.Abs(exePath)

	fmt.Printf("App - ReadConfigFile(): home: %s\n", homeDir)
	fmt.Printf("App - ReadConfigFile(): pwd: %s\n", currentDir)
	fmt.Printf("App - ReadConfigFile(): exe: %s\n", exePath)

	if err != nil {
		return config, err
	}

	exeDir := path.Dir(exePath)

	fmt.Printf("App - ReadConfigFile(): dir: %s\n", exeDir)

	configFile, err = findConfigFile(currentDir)

	fmt.Printf("App - ReadConfigFile(): config: %s; error: %#v\n", configFile, err)

	if err != nil {
		configFile, err = findConfigFile(exeDir)

		fmt.Printf("App - ReadConfigFile(): config: %s; error: %#v\n", configFile, err)
	}

	if err != nil {
		configFile, err = findConfigFile(homeDir)

		fmt.Printf("App - ReadConfigFile(): config: %s; error: %#v\n", configFile, err)
	}

	if configFile != "" && err == nil {
		var configData []byte

		// Load the file; returns []byte
		if configData, err = ioutil.ReadFile(configFile); err != nil {
			return config, err
		}

		// Unmarshal our input YAML file into empty Car (var c)
		if err = yaml.Unmarshal(configData, &config); err != nil {
			return config, err
		}

		config.MainDirectory = path.Dir(configFile)
		config.ConfigFile = configFile
	}

	return config, err
}
