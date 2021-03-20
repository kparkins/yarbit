package main

import "github.com/spf13/cobra"

func txCommand() *cobra.Command {
	command := &cobra.Command{
		Use: "tx",
		Short: "Interact with txs (add...)",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	command.AddCommand(txAddCommand())
	return command
}

func txAddCommand() *cobra.Command {

	return &cobra.Command{

	}
}

