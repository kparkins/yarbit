package main

import (
	"fmt"
	"github.com/kparkins/yarbit/database"
	"github.com/spf13/cobra"
	"os"
)

func balancesCommand() *cobra.Command {
	command := &cobra.Command{
		Use: "balances",
		Short: "Interact with balances (list...)",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	command.AddCommand(balancesListCommand())
	return command
}

func balancesListCommand() *cobra.Command {
	command := &cobra.Command{
		Use: "list",
		Short: "List all balances.",
		Run: func(cmd *cobra.Command, args []string) {
			dataDir, _ := cmd.Flags().GetString(flagDataDir)
			state := database.NewStateFromDisk(dataDir)
			if err := state.Load(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Printf("Account balances at %x\n", state.LatestBlockHash())
			fmt.Printf("------------------\n\n")
			for account, balance := range state.Balances {
				fmt.Println(fmt.Sprintf("%10s: %10d", account, balance))
			}
		},
	}
	addDefaultRequiredFlags(command)
	return command
}