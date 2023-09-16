package backup

import (
	"errors"

	log "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Perform backups & related operations",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		log.Error().Err(errors.New("test error")).Msg("Printing help")
		if err := cmd.Help(); err != nil {
			log.Error().Err(err).Msg("error printing help")
		}
	},
}

func init() {
	BackupCmd.AddCommand(addCmd)
	BackupCmd.AddCommand(purgeCmd)
	BackupCmd.AddCommand(listCmd)
}
