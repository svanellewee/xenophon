package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/svanellewee/xenophon/storage"
	"github.com/svanellewee/xenophon/storage/engines/sqlite3"
)

var configFile string

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

const (
	engineKey       = "storageengine"
	databaseFileKey = "databasepath"
	configName      = "config"
	configType      = "yaml"
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

func makeConfigFile(configHome string) error {

	viper.AddConfigPath(configHome)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	viper.SetDefault(databaseFileKey, filepath.Join(configHome, "history.db"))
	viper.SetDefault(engineKey, "sqlite3")

	configFile = filepath.Join(configHome, configName+"."+configType)
	if _, err := os.Stat(configFile); err != nil {
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

var (
	database *storage.DatabaseModule
)

// initEngine creates the backend specified by the `engineKey`
func initEngine() {
	engine := viper.GetString(engineKey)
	switch engine {
	// case "memory":
	// 	db := memory.NewMemoryStore()
	// 	database = storage.NewStorageModule(db)
	default:
		WarningLogger.Printf("Unknown engine %s, default to sqlite3", engine)
		fallthrough
	case "sqlite3":
		db := sqlite3.NewSqliteStorage(viper.GetString(databaseFileKey))
		database = storage.NewStorageModule(db)
	}
}

func init() {
	cobra.OnInitialize(
		initLoggers,
		initConfig,
		initEngine)
}

func initConfig() {

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

var rootCmd = &cobra.Command{
	Use:   "xenophon",
	Short: "Xenophon is a drop-in replacement for your shell history",
	Long:  `Xenophon stores your bash history in a datastore. It supports multiple backends`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
