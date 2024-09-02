package config

import (
	"log"
	"log/slog"

	commonConfig "github.com/hibare/GoCommon/v2/pkg/config"
	commonLogger "github.com/hibare/GoCommon/v2/pkg/logger"
	commonUtils "github.com/hibare/GoCommon/v2/pkg/utils"
	"github.com/hibare/GoS3Backup/internal/constants"
)

type S3Config struct {
	Endpoint  string `yaml:"endpoint" mapstructure:"endpoint"`
	Region    string `yaml:"region" mapstructure:"region"`
	AccessKey string `yaml:"access-key" mapstructure:"access-key"`
	SecretKey string `yaml:"secret-key" mapstructure:"secret-key"`
	Bucket    string `yaml:"bucket" mapstructure:"bucket"`
	Prefix    string `yaml:"prefix" mapstructure:"prefix"`
}

type GPGConfig struct {
	KeyServer string `yaml:"key-server" mapstructure:"key-server"`
	KeyID     string `yaml:"key-id" mapstructure:"key-id"`
}

type Encryption struct {
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
	GPG     GPGConfig
}

type BackupConfig struct {
	Dirs           []string   `yaml:"dirs" mapstructure:"dirs"`
	Hostname       string     `yaml:"-"`
	RetentionCount int        `yaml:"retention-count" mapstructure:"retention-count"`
	DateTimeLayout string     `yaml:"date-time-layout" mapstructure:"date-time-layout"`
	Cron           string     `yaml:"cron" mapstructure:"cron"`
	ArchiveDirs    bool       `yaml:"archive-dirs" mapstructure:"archive-dirs"`
	Encryption     Encryption `yaml:"encryption" mapstructure:"encryption"`
}

type DiscordNotifierConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Webhook string `yaml:"webhook" mapstructure:"webhook"`
}

type NotifiersConfig struct {
	Enabled bool                  `yaml:"enabled" mapstructure:"enabled"`
	Discord DiscordNotifierConfig `yaml:"discord" mapstructure:"discord"`
}

type LoggerConfig struct {
	Level string `yaml:"level" mapstructure:"level"`
	Mode  string `yaml:"mode" mapstructure:"mode"`
}

type Config struct {
	S3        S3Config        `yaml:"s3" mapstructure:"s3"`
	Backup    BackupConfig    `yaml:"backup" mapstructure:"backup"`
	Notifiers NotifiersConfig `yaml:"notifiers" mapstructure:"notifiers"`
	Logger    LoggerConfig    `yaml:"logger" mapstructure:"logger"`
}

var Current *Config

var BC commonConfig.BaseConfig

func LoadConfig() {
	current, err := BC.ReadYAMLConfig(Current)
	if err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}
	Current = current.(*Config)

	// Check if logger.level & logger.mode are correct
	if Current.Logger.Level == "" {
		Current.Logger.Level = commonLogger.DefaultLoggerLevel
	} else if Current.Logger.Level != "" {
		if !commonLogger.IsValidLogLevel(Current.Logger.Level) {
			log.Fatalf("Error invalid logger level: %s", Current.Logger.Level)
		}
	}

	if Current.Logger.Mode == "" {
		Current.Logger.Mode = commonLogger.DefaultLoggerMode
	} else if Current.Logger.Mode != "" {
		if !commonLogger.IsValidLogMode(Current.Logger.Mode) {
			log.Fatalf("Error invalid logger mode: %s", Current.Logger.Mode)
		}
	}

	commonLogger.InitLogger(&Current.Logger.Level, &Current.Logger.Mode)

	// Set default DateTimeLayout if missing
	if Current.Backup.DateTimeLayout == "" {
		slog.Warn("DateTimeLayout is not set, using default", "default", constants.DefaultDateTimeLayout)
		Current.Backup.DateTimeLayout = constants.DefaultDateTimeLayout
	}

	// Set RetentionCount if missing
	if Current.Backup.RetentionCount == 0 {
		slog.Warn("RetentionCount is not set, using default", "default", constants.DefaultRetentionCount)
		Current.Backup.RetentionCount = constants.DefaultRetentionCount
	}

	// Set Schedule if missing
	if Current.Backup.Cron == "" {
		slog.Warn("Schedule is not set, using default", "default", constants.DefaultCron)
		Current.Backup.Cron = constants.DefaultCron
	}

	// If notifier webhook is empty, set status to disable
	if Current.Notifiers.Discord.Webhook == "" {
		Current.Notifiers.Discord.Enabled = false
	}

	// Check if encryption is enabled & encryption config is enabled
	if Current.Backup.Encryption.Enabled && !Current.Backup.ArchiveDirs {
		slog.Warn("Backup encryption is only available when archive dirs are enabled. Disabling encryption")
		Current.Backup.Encryption.Enabled = false
	} else if Current.Backup.Encryption.Enabled {
		if Current.Backup.Encryption.GPG.KeyServer == "" || Current.Backup.Encryption.GPG.KeyID == "" {
			slog.Error("Encryption is enabled but GPG key server or key ID is missing")
			Current.Backup.Encryption.Enabled = false
		}
	}

	Current.Backup.Hostname = commonUtils.GetHostname()
}

func CleanConfig() error {
	return BC.CleanConfig()
}

func InitConfig() error {
	if Current == nil {
		Current = &Config{}
	}

	if err := BC.WriteYAMLConfig(Current); err != nil {
		return err
	}

	return nil
}

func init() {
	BC = commonConfig.BaseConfig{
		ProgramIdentifier: constants.ProgramIdentifier,
		OS:                commonConfig.ActualOS{},
	}
	if err := BC.Init(); err != nil {
		panic(err)
	}
}
