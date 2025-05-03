package cmd

import (
	"fmt"
	"os"

	"github.com/gkwa/sunlitsparrow/internal/logger"
	"github.com/spf13/cobra"
)

var verbosity int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sunlitsparrow",
	Short: "Explore and query Maccy's clipboard history database",
	Long:  `A tool to explore and query the SQLite database used by Maccy to store clipboard history.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.SetLogLevel(verbosity)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "increase verbosity level")

	// Add subcommands
	rootCmd.AddCommand(schemaCmd)
	rootCmd.AddCommand(itemsCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(pinsCmd)
}
