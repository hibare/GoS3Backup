package backup

import (
	"github.com/hibare/GoS3Backup/internal/backup"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Perform a backup",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		backup.Backup()
	},
}
