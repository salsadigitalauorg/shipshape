package cmd

import (
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage shipshape plugins",
	Run: func(cmd *cobra.Command, args []string) {
		plugin.Run()
	},
}

func init() {
	rootCmd.AddCommand(pluginCmd)
}
