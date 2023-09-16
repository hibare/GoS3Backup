package cmd

import (
	"os"
	"time"

	"github.com/go-co-op/gocron"
	commonLogger "github.com/hibare/GoCommon/v2/pkg/logger"
	"github.com/hibare/GoS3Backup/cmd/backup"
	configCmd "github.com/hibare/GoS3Backup/cmd/config"
	intBackup "github.com/hibare/GoS3Backup/internal/backup"
	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/hibare/GoS3Backup/internal/constants"
	"github.com/hibare/GoS3Backup/internal/version"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "GoS3Backup",
	Short:   "Application to backup directories to S3",
	Long:    "",
	Version: version.CurrentVersion,
	Run: func(cmd *cobra.Command, args []string) {

		s := gocron.NewScheduler(time.UTC)

		// Schedule backup job
		if _, err := s.Cron(config.Current.Backup.Cron).Do(func() {
			intBackup.Backup()
			intBackup.PurgeOldBackups()
		}); err != nil {
			log.Error().Err(err).Msg("Error setting up cron")
		}
		log.Info().Msgf("Scheduled backup job to run every %s", config.Current.Backup.Cron)

		// Schedule version check job
		if _, err := s.Cron(constants.VersioCheckCron).Do(func() {
			version.V.CheckUpdate()
		}); err != nil {
			log.Warn().Err(err).Msg("Failed to schedule version check job")
		}

		s.StartBlocking()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(configCmd.ConfigCmd)
	rootCmd.AddCommand(backup.BackupCmd)

	cobra.OnInitialize(commonLogger.InitLogger, config.LoadConfig)

	initialVersionCheck := func() {
		version.V.CheckUpdate()
		if version.V.NewVersionAvailable {
			log.Info().Msg(version.V.GetUpdateNotification())
		}
	}
	go initialVersionCheck()
}
