/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package init

import (
	"fmt"

	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize application (create necessary config and other things)",
	Run: func(cmd *cobra.Command, args []string) {
		config.LoadConfig()
		configRootDir := config.GetConfigRootDir()
		configFilePath := config.GetConfigFilePath(configRootDir)
		fmt.Printf("\n\nConfig file path: %s\n", configFilePath)
		fmt.Printf("Empty config file is loaded at above location. Edit config as per your needs.\n\n")
	},
}
