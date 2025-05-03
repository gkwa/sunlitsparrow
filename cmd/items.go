package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/gkwa/sunlitsparrow/internal/db"
	"github.com/gkwa/sunlitsparrow/internal/history"
	"github.com/spf13/cobra"
)

var (
	itemsTableFormat bool
	itemsLimit       int
)

// itemsCmd represents the items command
var itemsCmd = &cobra.Command{
	Use:   "items",
	Short: "List clipboard items",
	Run: func(cmd *cobra.Command, args []string) {
		dbConn, err := db.OpenMaccyDB()
		if err != nil {
			cmd.PrintErrln("Error opening database:", err)
			return
		}
		defer dbConn.Close()

		historyRepo := history.NewRepository(dbConn)
		items, err := historyRepo.GetRecentItems(itemsLimit)
		if err != nil {
			cmd.PrintErrln("Error retrieving items:", err)
			return
		}

		if len(items) == 0 {
			cmd.Println("No items found.")
			return
		}

		if itemsTableFormat {
			// Print in table format
			printer := history.NewPrinter(items)
			printer.PrintItems()
		} else {
			// Print in JSON format (default)
			jsonData, err := json.MarshalIndent(items, "", "  ")
			if err != nil {
				cmd.PrintErrln("Error formatting JSON:", err)
				return
			}
			fmt.Println(string(jsonData))
		}
	},
}

func init() {
	itemsCmd.Flags().BoolVarP(&itemsTableFormat, "table", "t", false, "Display output in table format instead of JSON")
	itemsCmd.Flags().IntVarP(&itemsLimit, "limit", "l", 10, "Limit the number of items to display")
}
