package config

import (
	"fmt"

	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/spf13/cobra"
)

var InitConfigCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize application config",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.InitConfig(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("\n\nConfig file path: %s\n", config.BC.ConfigFilePath)
			fmt.Printf("Empty config file is loaded at above location. Edit config as per your needs.\n\n")
		}
	},
}
