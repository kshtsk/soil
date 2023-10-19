package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"soil/deploy"
)

func init() {
	rootCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:     "remove [NAME]",
	Aliases: []string{"rm"},
	Short:   "Remove deployment with given name",
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
		fmt.Printf("Removing %s...\n", d.DName())
		d.Remove(Force)
	},
}
