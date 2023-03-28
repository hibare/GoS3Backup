package cmd

import (
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/hibare/GoS3Backup/cmd/backup"
	initialize "github.com/hibare/GoS3Backup/cmd/init"
	backup_int "github.com/hibare/GoS3Backup/internal/backup"
	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/hibare/GoS3Backup/internal/logging"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:     "GoS3Backup",
	Short:   "Application to backup directories to S3",
	Long:    "",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		s := gocron.NewScheduler(time.UTC)
		if _, err := s.Cron(config.Current.Backup.Cron).Do(func() {
			backup_int.Backup()
			backup_int.PurgeOldBackups()
		}); err != nil {
			log.Fatalf("Error cron: %v", err)
		}
		log.Infof("Scheduled backup job to run every %s", config.Current.Backup.Cron)
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
	rootCmd.AddCommand(backup.BackupCmd)

	cobra.OnInitialize(logging.SetupLogger, config.LoadConfig)
}
