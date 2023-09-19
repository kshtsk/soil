package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"soil/deploy"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:     "status [NAME]",
	Aliases: []string{"st", "stat", "state"},
	Short:   "Print deployment status",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			d, err := deploy.LookupDeployment(args[0])
			if err != nil {
				log.Fatalf("Deployment '%s' does not exists", args[0])
			}
			fmt.Printf("Deployment %s\n", d)
		} else {
			dd := deploy.LookupDeployments()
			if len(dd) < 1 {
				fmt.Printf("No deployments present.\n")
				return
			}
			for _, d := range dd {
				/*
					t := ""
					switch d.(type) {
					case *deploy.ScalabilityDeployment:
						t = d.(*deploy.ScalabilityDeployment).Kind
					default:
						log.Printf("WARNING: Unhandled type: s", reflect.TypeOf(d).Elem())
						t = "unknown"
					}
				*/
				fmt.Printf("- %s\n", d.Brief())
			}
		}
	},
}
