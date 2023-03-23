/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package backup

import (
	"github.com/hibare/GoS3Backup/internal/backup"
	"github.com/spf13/cobra"
)

// purgeCmd represents the purge command
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge old backups",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		backup.PurgeOldBackups()
	},
}
