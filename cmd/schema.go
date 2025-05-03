package cmd

import (
	"github.com/gkwa/sunlitsparrow/internal/db"
	"github.com/gkwa/sunlitsparrow/internal/schema"
	"github.com/spf13/cobra"
)

var outputFile string

// schemaCmd represents the schema command
var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Show database schema",
	Run: func(cmd *cobra.Command, args []string) {
		dbConn, err := db.OpenMaccyDB()
		if err != nil {
			cmd.PrintErrln("Error opening database:", err)
			return
		}
		defer dbConn.Close()

		explorer := schema.NewExplorer(dbConn)

		if outputFile != "" {
			if err := explorer.ExportSchemaToFile(outputFile); err != nil {
				cmd.PrintErrln("Error exporting schema to file:", err)
				return
			}
			cmd.Printf("Schema exported to %s\n", outputFile)
			return
		}

		explorer.ExploreSchema()
	},
}

func init() {
	schemaCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output schema to SQLite-compatible file")
}
