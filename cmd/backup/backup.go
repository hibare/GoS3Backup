package backup

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Perform backups & related operations",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			log.Fatalf("Error: %v", err)
		}
	},
}

func init() {
	BackupCmd.AddCommand(addCmd)
	BackupCmd.AddCommand(purgeCmd)
	BackupCmd.AddCommand(listCmd)
}
