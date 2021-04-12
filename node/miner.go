package node

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kparkins/yarbit/database"
)

func displayMiningProgress(attempt int32) {
	if attempt%1000000 == 0 {
		fmt.Printf("mining attempt %d\n", attempt)
	}
}

type Miner struct {
	newBlockChan chan *database.Block
}

func NewMiner(blockChan chan *database.Block) *Miner {
	return &Miner{
		newBlockChan: blockChan,
	}
}

func (m *Miner) Start(ctx context.Context) {
	mining := false
	cancelMiner := func() {}
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			cancelMiner()
			return
		case <-m.newBlockChan:
			cancelMiner()
			mining = false
			// TODO
			/*hash, err := n.AddBlock(block)
			if err != nil {
				fmt.Printf("error adding new block %s\n", hash.String())
				break
			}
			if err := n.CompleteTxs(block.Txs); err != nil {
				fmt.Println(err)
				break
			}*/
			mining, cancelMiner = m.launch(ctx)
		case <-ticker.C:
			if mining {
				break
			}
			mining, cancelMiner = m.launch(ctx)
		}
	}
}

func (m *Miner) launch(ctx context.Context) (bool, context.CancelFunc) {
	pendingBlock := n.createPendingBlock()
	if len(pendingBlock.Txs) <= 0 {
		return false, func() {}
	}
	c, cancelMiner := context.WithCancel(ctx)
	go mine(c, pendingBlock, m.newBlockChan)
	return true, cancelMiner
}

func mine(ctx context.Context, pending *database.Block, minedBlock chan<- *database.Block) {
	if len(pending.Txs) <= 0 {
		fmt.Fprintln(os.Stderr, "cannot mine an empty block")
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