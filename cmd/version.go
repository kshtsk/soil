package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
func Version() string {
	return fmt.Sprintf("v%s %s/%s", "1.0", runtime.GOOS, runtime.GOARCH)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", Version())
	},
}
