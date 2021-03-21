package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

const FixVersion = "0"
const MinorVersion = "4"
const MajorVersion = "0"
const Description = "The Yarbit Ledger - Node Status"

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		Short: "Print the version of the Yarbit CLI.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s.%s.%s - %s\n", MajorVersion, MinorVersion, FixVersion, Description)
		},
	}
}
