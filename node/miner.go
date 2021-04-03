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
func (n *Node) startForeman(ctx context.Context) {
	mining := false
	cancelMiner := func() {}
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			cancelMiner()
			return
		case block := <-n.newBlockChan:
			cancelMiner()
			mining = false
			hash, err := n.AddBlock(block)
			if err != nil {
				fmt.Printf("error adding new block %s\n", hash.String())
				break
			}
			if err := n.CompleteTxs(block.Txs); err != nil {
				fmt.Println(err)
				break
			}
			mining, cancelMiner = n.startMiner(ctx, n.newBlockChan)
		case <-ticker.C:
			if mining {
				break
			}
			mining, cancelMiner = n.startMiner(ctx, n.newBlockChan)
		}
	}
}

func (n *Node) startMiner(ctx context.Context, minedBlockChan chan<- *database.Block) (bool, context.CancelFunc) {
	pendingBlock := n.createPendingBlock()
	if len(pendingBlock.Txs) <= 0 {
		return false, func() {}
	}
	c, cancelMiner := context.WithCancel(ctx)
	go mine(c, pendingBlock, minedBlockChan)
	return true, cancelMiner
}

func (n *Node) createPendingBlock() *database.Block {
	n.lock.RLock()
	defer n.lock.RUnlock()
	txs := make([]database.Tx, 0, len(n.pendingTxs))
	for _, tx := range n.pendingTxs {
		txs = append(txs, tx)
	}
	return &database.Block{
		Header: database.BlockHeader{
			Parent: n.state.LatestBlockHash(),
			Number: n.state.NextBlockNumber(),
			Nonce:  0,
			Time:   uint64(time.Now().Unix()),
			Miner:  n.config.MinerAccount,
		},
		Txs: txs,
	}
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
