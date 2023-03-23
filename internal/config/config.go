package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hibare/GoS3Backup/internal/constants"
	"github.com/hibare/GoS3Backup/internal/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type S3Config struct {
	Endpoint  string `yaml:"endpoint"`
	Region    string `yaml:"region"`
	AccessKey string `yaml:"access-key"`
	SecretKey string `yaml:"secret-key"`
	Bucket    string `yaml:"bucket"`
	Prefix    string `yaml:"prefix"`
}

type BackupConfig struct {
	Dirs           []string `yaml:"dirs"`
	Hostname       string
	RetentionCount int    `yaml:"retention-count"`
	DateTimeLayout string `yaml:"date-time-layout"`
	Cron           string `yaml:"cron"`
}

type DiscordNotifierConfig struct {
	Enabled bool   `yaml:"enabled"`
	Webhook string `yaml:"webhook"`
}

type NotifiersConfig struct {
	Enabled bool                  `yaml:"enabled"`
	Discord DiscordNotifierConfig `yaml:"discord"`
}

type Config struct {
	S3        S3Config        `yaml:"s3"`
	Backup    BackupConfig    `yaml:"backup"`
	Notifiers NotifiersConfig `yaml:"notifiers"`
}

var Current *Config

func LoadConfig() {
	configRootDir := getConfigRootDir()
	configFilePath := getConfigFilePath(configRootDir)
	preCheckConfigPath(configRootDir)

	file, err := os.Open(configFilePath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Error reading YAML data: %v", err)
	}

	// Unmarshal YAML data into struct
	if err := yaml.Unmarshal(data, &Current); err != nil {
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

func getConfigRootDir() string {
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

func getConfigFilePath(configRootDir string) string {
	return filepath.Join(configRootDir, fmt.Sprintf("%s.%s", constants.ConfigFilename, constants.ConfigFileExtension))
}

func preCheckConfigPath(configRootDir string) {

	configPath := getConfigFilePath(configRootDir)

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
