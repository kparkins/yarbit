package node

import (
	"context"
	"fmt"
	"github.com/kparkins/yarbit/database"
	"os"
	"time"
)

func mine(ctx context.Context, pending *database.Block, minedBlock chan *database.Block) {
	if len(pending.Txs) <= 0 {
		fmt.Fprintln(os.Stderr, "cannot mine an empty block")
		return
	}
	attempt := 0
	start := time.Now()
	hash, err := pending.Hash()
	fmt.Println(database.IsBlockHashValid(hash))
	for !database.IsBlockHashValid(hash) {
		select {
		case <-ctx.Done():
			fmt.Println("mining cancelled")
			return
		default:
			if attempt%100000 == 0 {
				fmt.Printf("mining attempt %d\n", attempt)
			}
			hash, err = pending.Hash()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error hashing new pending block")
				return
			}
			attempt++
			pending.Header.Nonce++
		}
	}

	fmt.Printf("\nmined new Block '%x'\n", hash)
	fmt.Printf("\theight: '%v'\n", pending.Header.Number)
	fmt.Printf("\tnonce: '%v'\n", pending.Header.Nonce)
	fmt.Printf("\tcreated: '%v'\n", pending.Header.Time)
	fmt.Printf("\tminer: '%v'\n", pending.Header.Miner)
	fmt.Printf("\tparent: '%s'\n\n", pending.Header.Parent.String())

	fmt.Printf("\tAttempt: '%v'\n", attempt)
	fmt.Printf("\tTime: %s\n\n", time.Since(start))
	minedBlock <- pending
}
