package node

import (
	"fmt"
	"sync"
	"time"

	"github.com/kparkins/yarbit/database"
)

type TxPool struct {
	lock         *sync.RWMutex
	pendingTxs   map[database.Hash]database.Tx
	completedTxs map[database.Hash]database.Tx // TODO need to expire or write to disk periodically
}

func NewTxPool() TxPool {
	return TxPool{
		lock:         &sync.RWMutex{},
		pendingTxs:   make(map[database.Hash]database.Tx),
		completedTxs: make(map[database.Hash]database.Tx),
	}
}

func (t *TxPool) RemoveTxs(txs []database.Tx) {
	for _, tx := range txs {
		hash, err := tx.Hash()
		if err != nil {
			fmt.Println(err)
			continue
		}
		delete(t.pendingTxs, hash)
	}
}

func (t *TxPool) CompleteTxs(txs []database.Tx) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, tx := range txs {
		hash, err := tx.Hash()
		if err != nil {
			return err
		}
		t.completedTxs[hash] = tx
		delete(t.pendingTxs, hash)
	}
	return nil
}

func (t *TxPool) AddPendingTx(tx database.Tx) (database.Hash, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	var hash database.Hash
	hash, err := tx.Hash()
	if err != nil {
		fmt.Printf("error hashing new tx %v\n", tx)
		return hash, err
	}
	if _, ok := t.completedTxs[hash]; ok {
		return hash, nil
	}
	t.pendingTxs[hash] = tx
	return hash, nil
}

func (t *TxPool) AddPendingTxs(txs []database.Tx) error {
	for _, tx := range txs {
		t.AddPendingTx(tx)
	}
	return nil
}

func (t *TxPool) PendingTxs() []database.Tx {
	t.lock.RLock()
	defer t.lock.RUnlock()
	txs := make([]database.Tx, 0, len(t.pendingTxs))
	for _, v := range t.pendingTxs {
		txs = append(txs, v)
	}
	return txs
}

func (t *TxPool) createPendingBlock() *database.Block {
	t.lock.RLock()
	defer t.lock.RUnlock()
	txs := make([]database.Tx, 0, len(t.pendingTxs))
	for _, tx := range t.pendingTxs {
		txs = append(txs, tx)
	}
	return &database.Block{
		Header: database.BlockHeader{
			Nonce: 0,
			Time:  uint64(time.Now().Unix()),
		},
		Txs: txs,
	}
}
