package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"soil/deploy"
)

func init() {
	rootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:     "deploy [NAME]",
	Aliases: []string{"dp"},
	Short:   "Create deployment under given name",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		kind := DeploymentType
		name := "default"
		if len(args) > 0 {
			name = args[0]
		}
		fmt.Printf("Deploying %s as %s...\n", kind, name)
		k := deploy.KindMap[kind]
		if deploy.TerraformVarFile != "" {
			k.TerraformVarFile = deploy.TerraformVarFile
		}
		if deploy.TerraformWorkDir != "" {
			k.TerraformWorkDir = deploy.TerraformWorkDir
		}
		k.TerraformRepoRef = deploy.TerraformRepoRef
		d := deploy.MakeDeployment(name, k)
		fmt.Printf("Created %s\n", d.Make())
	},
}
