package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

const FixVersion = "0"
const MinorVersion = "1"
const MajorVersion = "0"
const Description = "The Yarbit Ledger"

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		Short: "Print the version of the Yarbit CLI.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s.%s.%s - %s", MajorVersion, MinorVersion, FixVersion, Description)
		},
	}
}