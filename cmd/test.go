package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"soil/deploy"
)

func init() {
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:     "test [NAME]",
	Aliases: []string{"ts"},
	Short:   "Run test on the given name",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := "default"
		if len(args) > 0 {
			name = args[0]
		}
		d, err := deploy.LookupDeployment(name)
		if err != nil {
			log.Fatalf("No deployment found with name '%s': %v", name, err)
		}
		d.Test()
	},
}
