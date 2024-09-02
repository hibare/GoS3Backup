package backup

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Perform backups & related operations",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			slog.Error("error printing help")
		}
	},
}

func init() {
	BackupCmd.AddCommand(addCmd)
	BackupCmd.AddCommand(purgeCmd)
	BackupCmd.AddCommand(listCmd)
}
