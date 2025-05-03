package cmd

import (
	"github.com/gkwa/sunlitsparrow/internal/db"
	"github.com/gkwa/sunlitsparrow/internal/export"
	"github.com/gkwa/sunlitsparrow/internal/history"
	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [file]",
	Short: "Export clipboard items to JSON",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbConn, err := db.OpenMaccyDB()
		if err != nil {
			cmd.PrintErrln("Error opening database:", err)
			return
		}
		defer dbConn.Close()

		outputFile := "maccy-export.json"
		if len(args) > 0 {
			outputFile = args[0]
		}

		historyRepo := history.NewRepository(dbConn)
		items, err := historyRepo.GetAllItems()
		if err != nil {
			cmd.PrintErrln("Error retrieving items:", err)
			return
		}

		exporter := export.NewJSONExporter(outputFile)
		if err := exporter.Export(items); err != nil {
			cmd.PrintErrln("Error exporting items:", err)
			return
		}

		cmd.Printf("Exported %d items to %s\n", len(items), outputFile)
	},
}
