package cmd

import (
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/hibare/GoS3Backup/cmd/backup"
	"github.com/hibare/GoS3Backup/cmd/clean"
	initialize "github.com/hibare/GoS3Backup/cmd/init"
	backup_int "github.com/hibare/GoS3Backup/internal/backup"
	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/hibare/GoS3Backup/internal/constants"
	"github.com/hibare/GoS3Backup/internal/logging"
	"github.com/hibare/GoS3Backup/internal/version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "GoS3Backup",
	Short:   "Application to backup directories to S3",
	Long:    "",
	Version: version.CurrentVersion,
	Run: func(cmd *cobra.Command, args []string) {
		go version.CheckLatestRelease()

		s := gocron.NewScheduler(time.UTC)

		// Schedule backup job
		if _, err := s.Cron(config.Current.Backup.Cron).Do(func() {
			backup_int.Backup()
			backup_int.PurgeOldBackups()
		}); err != nil {
			log.Fatalf("Error cron: %v", err)
		}
		log.Infof("Scheduled backup job to run every %s", config.Current.Backup.Cron)

		// Schedule version check job
		if _, err := s.Cron(constants.VersioCheckCron).Do(func() {
			version.CheckLatestRelease()
		}); err != nil {
			log.Warnf("Failed to scedule version check job: %v", err)
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
	rootCmd.AddCommand(initialize.InitCmd)
	rootCmd.AddCommand(clean.CleanCmd)
	rootCmd.AddCommand(backup.BackupCmd)

	cobra.OnInitialize(logging.SetupLogger, config.LoadConfig)
}
