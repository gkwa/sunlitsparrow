package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/gkwa/sunlitsparrow/internal/db"
	"github.com/gkwa/sunlitsparrow/internal/history"
	"github.com/spf13/cobra"
)

var tableFormat bool

// pinsCmd represents the pins command
var pinsCmd = &cobra.Command{
	Use:   "pins",
	Short: "List pinned clipboard items",
	Run: func(cmd *cobra.Command, args []string) {
		dbConn, err := db.OpenMaccyDB()
		if err != nil {
			cmd.PrintErrln("Error opening database:", err)
			return
		}
		defer dbConn.Close()

		historyRepo := history.NewRepository(dbConn)
		pinnedItems, err := historyRepo.GetPinnedItems()
		if err != nil {
			cmd.PrintErrln("Error retrieving pinned items:", err)
			return
		}

		if len(pinnedItems) == 0 {
			cmd.Println("No pinned items found.")
			return
		}

		if tableFormat {
			// Print in table format
			printer := history.NewPrinter(pinnedItems)
			printer.PrintItems()
		} else {
			// Print in JSON format (default)
			jsonData, err := json.MarshalIndent(pinnedItems, "", "  ")
			if err != nil {
				cmd.PrintErrln("Error formatting JSON:", err)
				return
			}
			fmt.Println(string(jsonData))
		}
	},
}

func init() {
	pinsCmd.Flags().BoolVarP(&tableFormat, "table", "t", false, "Display output in table format instead of JSON")
}
