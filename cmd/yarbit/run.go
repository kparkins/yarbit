package main

import (
	"fmt"
	"github.com/kparkins/yarbit/node"
	"github.com/spf13/cobra"
	"os"
)

func runCommand() *cobra.Command {
	command := &cobra.Command{
		Use: "run",
		Short: "Launches the Yarbit node and its HTTP API.",
		Run: func(cmd *cobra.Command, args []string) {
			dataDir, err := cmd.Flags().GetString(flagDataDir)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			err = node.Run(dataDir)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

		},
	}
	addDefaultRequiredFlags(command)
	return command
}
