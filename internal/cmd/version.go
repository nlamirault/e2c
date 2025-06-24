package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nlamirault/e2c/internal/utils"
	"github.com/nlamirault/e2c/internal/version"

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s version %s\n", utils.APP_NAME, version.GetVersion())
		},
	}
}
