package clean

import (
	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/spf13/cobra"
)

var CleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Command to clean up program config",
	Run: func(cmd *cobra.Command, args []string) {
		config.CleanConfig()
	},
}
