package main

import (
	"fmt"
	"github.com/kparkins/yarbit/database"
	"github.com/spf13/cobra"
	"os"
)

const flagTo = "to"
const flagData = "data"
const flagFrom = "from"
const flagValue = "value"

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
	command := &cobra.Command{
		Use: "add",
		Short: "Add a transaction to the ledger.",
		Run: func(cmd *cobra.Command, args []string) {
			dataDir, _ := cmd.Flags().GetString(flagDataDir)
			from, _ := cmd.Flags().GetString(flagFrom)
			to, _ := cmd.Flags().GetString(flagTo)
			value, _:= cmd.Flags().GetUint(flagValue)
			data, _ := cmd.Flags().GetString(flagData)
			state := database.NewStateFromDisk(dataDir)
			if err := state.Load(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			_ = database.NewTx(database.NewAccount(from), database.NewAccount(to), value, data)
			//TODO
			/*if err := state.AddTx(tx); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if _, err := state.Persist(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}*/
			fmt.Println("TX successfully persisted to the ledger.")
		},
	}
	addDefaultRequiredFlags(command)

	command.Flags().String(flagFrom, "", "From what account to send tokens.")
	command.MarkFlagRequired(flagFrom)

	command.Flags().String(flagTo, "", "To what account to send tokens.")
	command.MarkFlagRequired(flagTo)

	command.Flags().Uint(flagValue, 0, "The amount of tokens to send.")
	command.MarkFlagRequired(flagValue)

	command.Flags().String(flagData, "", "Data to send with the transaction. 'reward' current used.")

	return command
}

