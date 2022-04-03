package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(insertCmd)
	//	rootCmd.AddCommand(listCmd)

}

var insertCmd = &cobra.Command{
	Use:   "insert",
	Short: "insert into history",
	//Long:  `Xenophon stores your bash history in a datastore. It supports multiple backends`,
	RunE: func(cmd *cobra.Command, args []string) error {
		database.Insert(args[0])
		return nil
	},
	// Run: func(cmd *cobra.Command, args []string) {
	// 	// Do Stuff Here
	// },
}
