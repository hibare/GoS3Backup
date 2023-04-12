package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hibare/GoS3Backup/internal/constants"
	"github.com/hibare/GoS3Backup/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type S3Config struct {
	Endpoint  string `yaml:"endpoint" mapstructure:"endpoint"`
	Region    string `yaml:"region" mapstructure:"region"`
	AccessKey string `yaml:"access-key" mapstructure:"access-key"`
	SecretKey string `yaml:"secret-key" mapstructure:"secret-key"`
	Bucket    string `yaml:"bucket" mapstructure:"bucket"`
	Prefix    string `yaml:"prefix" mapstructure:"prefix"`
}

type BackupConfig struct {
	Dirs           []string `yaml:"dirs" mapstructure:"dirs"`
	Hostname       string   `yaml:"-"`
	RetentionCount int      `yaml:"retention-count" mapstructure:"retention-count"`
	DateTimeLayout string   `yaml:"date-time-layout" mapstructure:"date-time-layout"`
	Cron           string   `yaml:"cron" mapstructure:"cron"`
	ZipDirs        bool     `yaml:"zip-dirs" mapstructure:"zip-dirs"`
}

type DiscordNotifierConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Webhook string `yaml:"webhook" mapstructure:"webhook"`
}

type NotifiersConfig struct {
	Enabled bool                  `yaml:"enabled" mapstructure:"enabled"`
	Discord DiscordNotifierConfig `yaml:"discord" mapstructure:"discord"`
}

type Config struct {
	S3        S3Config        `yaml:"s3" mapstructure:"s3"`
	Backup    BackupConfig    `yaml:"backup" mapstructure:"backup"`
	Notifiers NotifiersConfig `yaml:"notifiers" mapstructure:"notifiers"`
}

var Current *Config

func LoadConfig() {
	configRootDir := GetConfigRootDir()
	preCheckConfigPath(configRootDir)

	viper.SetConfigName(constants.ConfigFilename)
	viper.AddConfigPath(configRootDir)
	viper.SetConfigType(constants.ConfigFileExtension)

	// Load the configuration file
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Unmarshal the configuration file into a struct
	if err := viper.Unmarshal(&Current); err != nil {
		log.Fatalf("Error parsing YAML data: %v", err)
	}

	// Set default DateTimeLayout if missing
	if Current.Backup.DateTimeLayout == "" {
		log.Warnf("DateTimeLayout is not set, using default: %s", constants.DefaultDateTimeLayout)
		Current.Backup.DateTimeLayout = constants.DefaultDateTimeLayout
	}

	// Set RetentionCount if missing
	if Current.Backup.RetentionCount == 0 {
		log.Warnf("RetentionCount is not set, using default: %d", constants.DefaultRetentionCount)
		Current.Backup.RetentionCount = constants.DefaultRetentionCount
	}

	// Set Schedule if missing
	if Current.Backup.Cron == "" {
		log.Warnf("Schedule is not set, using default: %s", constants.DefaultCron)
		Current.Backup.Cron = constants.DefaultCron
	}

	Current.Backup.Hostname = utils.GetHostname()
}

func CleanConfig() {
	configRootDir := GetConfigRootDir()
	if info, err := os.Stat(configRootDir); err != nil && os.IsExist(err) {
		log.Fatalf("Error %s", err)
	} else if !info.IsDir() {
		log.Fatalf("Config directory is not a directory: %s", configRootDir)
		return
	} else {
		if err := os.RemoveAll(configRootDir); err != nil {
			log.Fatalf("Error removing config directory: %v", err)
		}
	}
}

func GetConfigRootDir() string {
	var configRootDir string

	switch os := runtime.GOOS; os {
	case "linux":
		configRootDir = constants.ConfigRootLinux
	case "windows":
		configRootDir = constants.ConfigRootWindows
	case "darwin":
		configRootDir = constants.ConfigRootDarwin
	default:
		log.Fatalf("Unsupported operating system: %s", os)
	}

	return filepath.Join(configRootDir, strings.ToLower(constants.ProgramIdentifier))
}

func GetConfigFilePath(configRootDir string) string {
	return filepath.Join(configRootDir, fmt.Sprintf("%s.%s", constants.ConfigFilename, constants.ConfigFileExtension))
}

func preCheckConfigPath(configRootDir string) {

	configPath := GetConfigFilePath(configRootDir)

	if info, err := os.Stat(configRootDir); os.IsNotExist(err) {
		log.Warnf("Config directory does not exist, creating: %s", configRootDir)
		if err := os.MkdirAll(configRootDir, 0755); err != nil {
			log.Fatalf("Error creating config directory: %v", err)
			return
		}
	} else if !info.IsDir() {
		log.Fatalf("Config directory is not a directory: %s", configRootDir)
		return
	}

	if info, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Warnf("Config file does not exist, creating: %s", configPath)
		file, err := os.Create(configPath)
		if err != nil {
			log.Fatalf("Error creating config file: %v", err)
			return
		}
		defer file.Close()

		// Marshal empty config
		yamlBytes, err := yaml.Marshal(Config{})
		if err != nil {
			log.Fatalf("Error marshaling config: %v", err)
		}

		// Write the YAML output to a file
		if _, err := file.Write(yamlBytes); err != nil {
			log.Fatalf("Error writing config file: %v", err)
		}

	} else if info.IsDir() {
		log.Fatalf("Expected file, found directory: %s", configPath)
		return
	}
}
