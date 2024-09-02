package cmd

import (
	"log/slog"
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
			slog.Error("Error setting up cron")
		}
		slog.Info("Scheduled backup job", "cron", config.Current.Backup.Cron)

		// Schedule version check job
		if _, err := s.Cron(constants.VersionCheckCron).Do(func() {
			version.V.CheckUpdate()
		}); err != nil {
			slog.Warn("Failed to schedule version check job")
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

	cobra.OnInitialize(commonLogger.InitDefaultLogger, config.LoadConfig)

	initialVersionCheck := func() {
		version.V.CheckUpdate()
		if version.V.NewVersionAvailable {
			slog.Info(version.V.GetUpdateNotification())
		}
	}
	go initialVersionCheck()
}
