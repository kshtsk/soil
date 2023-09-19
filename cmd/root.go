package cmd

import (
	"fmt"
	"os"
	"soil/deploy"

	"github.com/spf13/cobra"
	_ "github.com/spf13/viper"
)

// var DeploymentName string
var DeploymentType string
var Force bool

var rootCmd = &cobra.Command{
	Use:   "so",
	Short: "Soil is a tool for deploying systems in different environments.",
	Long: "Deploy one or more clusters, run test, observe the results,\n" +
		"for details, please, refer https://github.com/kshtsk/soil/",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		fmt.Printf("Hello, I'm Soil (%s).\n\n", Version())
		fmt.Printf("Try:\n     so help\n")
	},
}

func init() {
	// rootCmd.PersistentFlags().StringVarP(&DeploymentName, "name", "n", "default", "Deployment name")
	rootCmd.PersistentFlags().BoolVarP(&Force, "force", "f", false, "Force")
	deployCmd.Flags().StringVarP(&DeploymentType, "type", "t", "k3d", "Deployment Type")
	deployCmd.Flags().StringVarP(&deploy.TerraformRepoRef, "terraform-repo-ref", "r",
		"https://github.com/moio/scalability-tests", "Terraform git repo ref")
	deployCmd.Flags().StringVarP(&deploy.TerraformWorkDir, "terraform-work-dir", "w", "", "Terraform work dir")
	deployCmd.Flags().StringVarP(&deploy.TerraformVarFile, "terraform-var-file", "v", "", "Terraform var file")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
