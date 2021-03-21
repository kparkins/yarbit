package main

import (
	"fmt"
	"github.com/kparkins/yarbit/database"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Must supply directory as first argument to the migrate command.")
		os.Exit(1)
	}
	dataDir := os.Args[1]
	state, err := database.NewStateFromDisk(dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading state from disk.")
		os.Exit(1)
	}
	defer state.Close()
	block0 := database.NewBlock(
		database.Hash{},
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
}
