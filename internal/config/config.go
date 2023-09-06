package config

import (
	commonConfig "github.com/hibare/GoCommon/v2/pkg/config"
	commonUtils "github.com/hibare/GoCommon/v2/pkg/utils"
	"github.com/hibare/GoS3Backup/internal/constants"
	log "github.com/sirupsen/logrus"
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
	Enabled bool `yaml:"encrypt" mapstructure:"encrypt"`
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

type Config struct {
	S3        S3Config        `yaml:"s3" mapstructure:"s3"`
	Backup    BackupConfig    `yaml:"backup" mapstructure:"backup"`
	Notifiers NotifiersConfig `yaml:"notifiers" mapstructure:"notifiers"`
}

var Current *Config

var BC commonConfig.BaseConfig

func LoadConfig() {
	current, err := BC.ReadYAMLConfig(Current)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	Current = current.(*Config)

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

	// If notifier webhook is empty, set status to disable
	if Current.Notifiers.Discord.Webhook == "" {
		Current.Notifiers.Discord.Enabled = false
	}

	// Check if encryption is enabled & encryption config is enabled
	if Current.Backup.Encryption.Enabled && !Current.Backup.ArchiveDirs {
		log.Warningf("Backup encryption is only available when archive dirs are enabled. Disabling encryption")
		Current.Backup.Encryption.Enabled = false
	} else if Current.Backup.Encryption.Enabled {
		if Current.Backup.Encryption.GPG.KeyServer == "" || Current.Backup.Encryption.GPG.KeyID == "" {
			log.Fatalf("Error backup encryption is enabled but encryption config is not set")
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
