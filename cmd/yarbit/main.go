package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

const flagDataDir = "datadir"

func main() {
	command := &cobra.Command{
		Use: "yarbit",
		Short: "The yarbit command.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(0)
		},
	}

	command.AddCommand(versionCommand())
	command.AddCommand(balancesCommand())
	command.AddCommand(runCommand())

	err := command.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultRequiredFlags(command *cobra.Command) {
	command.Flags().String(flagDataDir, "", "Path to the database directory.")
	command.MarkFlagRequired(flagDataDir)
}
