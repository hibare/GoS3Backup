package backup

import (
	"fmt"
	"os"

	"github.com/hibare/GoS3Backup/internal/backup"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List backups",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		backups, err := backup.ListBackups()
		if err != nil {
			panic(err)
		} else if len(backups) <= 0 {
			fmt.Println("No backups found")
		} else {
			fmt.Printf("\nTotal backups %d\n", len(backups))
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.SetColumnConfigs([]table.ColumnConfig{
				{
					Name:     "Backup Key",
					WidthMin: 20,
					WidthMax: 64,
				},
			})
			t.AppendHeader(table.Row{"#", "Backup Key"})

			for i, backup := range backups {

				t.AppendRow([]interface{}{i + 1, backup})
				t.AppendSeparator()
			}

			t.Render()
		}
	},
}
