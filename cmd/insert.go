package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(insertCmd)
}

var insertCmd = &cobra.Command{
	Use:   "insert",
	Short: "insert into history",
	Long:  `Insert into history store`,
	RunE: func(cmd *cobra.Command, args []string) error {

		_, err := database.Insert(args[0])
		if err != nil {
			ErrorLogger.Printf("can't insert %v\n", err)
			return err
		}
		defer database.Storage.Close()
		return nil
	},
}
