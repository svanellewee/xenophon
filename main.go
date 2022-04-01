package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"log"

	"github.com/spf13/viper"
	storage "github.com/svanellewee/xenophon/storage"
	sqlite3 "github.com/svanellewee/xenophon/storage/engines/sqlite3"
)

var configFile string

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func initLoggers() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// buildConfigDir defines (and creates) the config dir if not present.
func buildConfigDir() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine homedir: %w", err)
	}

	configHome := path.Join(homedir, ".xenophon")
	if _, err = os.Stat(configHome); err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(configHome, os.ModePerm); err != nil {
				return "", fmt.Errorf("could not create config dir: %w", err)
			} else {
				InfoLogger.Printf("created config directory at %s", configHome)
			}
		} else {
			return "", fmt.Errorf("an unexpected error occured: %w", err)
		}
	}
	return configHome, nil
}

const databaseKey = "databasepath"

func makeConfigFile(configHome string) error {
	const configName = "config"
	const configType = "yaml"

	viper.AddConfigPath(configHome)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.SetDefault(databaseKey, filepath.Join(configHome, "history.db"))

	configFile = filepath.Join(configHome, configName+"."+configType)
	if _, err = os.Stat(configFile); err != nil {
		if os.IsNotExist(err) {
			if _, err := os.Create(configFile); err != nil {
				return fmt.Errorf("could not create configfile: %w", err)
			} else {
				if err = viper.WriteConfig(); err != nil {
					return fmt.Errorf("could not write configfile: %w", err)
				}
				InfoLogger.Printf("created config file at %s", configHome)
			}
		} else {
			return fmt.Errorf("could not find configfile: %w", err)
		}
	}
	return nil
}

func init() {

	initLoggers()

	// if configDir don't exist {
	// 	mkdir
	// }
	configDir, err := buildConfigDir()
	if err != nil {
		ErrorLogger.Fatalf("Failed to create config dir: %v", err)
	}

	// if file don't exist  {
	// 	mk the file with defaults
	// }
	err = makeConfigFile(configDir)
	if err != nil {
		ErrorLogger.Fatalf("Failed to create config file: %v", err)
	}

	// load file
	err = viper.ReadInConfig()
	if err != nil {
		ErrorLogger.Fatalf("Failed to read config file: %v", err)
	}
}

var err error

func main() {
	db := sqlite3.NewSqliteStorage(viper.GetString(databaseKey))
	storage := storage.NewStorageModule(db)
	_ = storage
}
