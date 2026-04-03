package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print hush version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("hush %s\n", version)
		},
	}
}
