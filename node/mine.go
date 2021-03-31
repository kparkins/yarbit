package node

import (
	"context"
	"fmt"
	"github.com/kparkins/yarbit/database"
	"os"
	"time"
)

func displayMiningProgress(attempt int32) {
	if attempt%1000000 == 0 {
		fmt.Printf("mining attempt %d\n", attempt)
	}
}

func mine(ctx context.Context, pending *database.Block, minedBlock chan<- *database.Block) {
	if len(pending.Txs) <= 0 {
		fmt.Fprintln(os.Stderr, "cannot startForeman an empty block")
		return
	}
	var err error
	start := time.Now()
	hash := database.Hash{0xFF}
	for ; !database.IsBlockHashValid(hash); pending.Header.Nonce++ {
		select {
		case <-ctx.Done():
			fmt.Println("mining cancelled")
			return
		default:
			displayMiningProgress(pending.Header.Nonce)
			hash, err = pending.Hash()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error hashing new pending block")
				return
			}
		}
	}

	fmt.Printf("\n\tmined new Block '%s'\n", hash.String())
	fmt.Printf("\theight: '%v'\n", pending.Header.Number)
	fmt.Printf("\tnonce: '%v'\n", pending.Header.Nonce)
	fmt.Printf("\tcreated: '%v'\n", pending.Header.Time)
	fmt.Printf("\tminer: '%v'\n", pending.Header.Miner)
	fmt.Printf("\tparent: '%s'\n\n", pending.Header.Parent.String())

	fmt.Printf("\ttime: %s\n\n", time.Since(start))
	minedBlock <- pending
}
