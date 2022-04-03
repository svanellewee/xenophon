package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list entire command history in current directory",
	Long:  `List entire command history in current directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		location, err := os.Getwd()
		if err != nil {
			ErrorLogger.Fatalf("could not determine location: %v", err)
		}
		for _, e := range database.Location(location).Output() {
			fmt.Println(e)
		}
		return nil
	},
}
