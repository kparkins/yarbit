package main

import (
	"fmt"
	"github.com/kparkins/yarbit/database"
	"github.com/spf13/cobra"
	"os"
	"time"
)

func migrateCommand() *cobra.Command {
	command := &cobra.Command{
		Use: "migrate",
		Short: "Run the database migration",
		Run: func(cmd *cobra.Command, args []string) {
			dataDir, _ := cmd.Flags().GetString(flagDataDir)
			state, err := database.NewStateFromDisk(dataDir)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error loading state from disk.")
				os.Exit(1)
			}
			defer state.Close()
			block0 := database.NewBlock(
				database.Hash{},
				0,
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("andrej", "andrej", 3, ""),
					database.NewTx("andrej", "andrej", 700, "reward"),
				},
			)

			state.AddBlock(block0)
			block0Hash, _ := state.Persist()

			block1 := database.NewBlock(
				block0Hash,
				block0.Header.Number + 1,
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("andrej", "babayaga", 2000, ""),
					database.NewTx("andrej", "andrej", 100, "reward"),
					database.NewTx("babayaga", "andrej", 1, ""),
					database.NewTx("babayaga", "caesar", 1000, ""),
					database.NewTx("babayaga", "andrej", 50, ""),
					database.NewTx("andrej", "andrej", 600, "reward"),
				},
			)

			state.AddBlock(block1)
			state.Persist()
		},
	}

	addDefaultRequiredFlags(command)
	return command
}
