package database

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
)

const MaxBlocksPerRead = 100

type State struct {
	balances      map[Account]uint
	dataDir       string
	txMempool     []Tx
	blockStore    BlockStore
	lastBlockHash Hash
	lastBlock     Block
}

func NewStateFromDisk(dataDir string) *State {
	blockDbPath := getBlockDatabaseFilePath(dataDir)
	state := &State{
		dataDir:       dataDir,
		balances:      make(map[Account]uint, 0),
		txMempool:     make([]Tx, 0),
		blockStore:    NewFileBlockStore(blockDbPath),
		lastBlockHash: Hash{},
		lastBlock:     NewBlock(Hash{}, 0, 0, make([]Tx, 0)),
	}
	return state
}

func (s *State) Balances() map[Account]uint {
	result := make(map[Account]uint, len(s.balances))
	for k, v := range s.balances {
		result[k] = v
	}
	return result
}

func (s *State) Load() error {
	if err := initDataDir(s.dataDir); err != nil {
		return err
	}
	genesis, err := LoadGenesis(getGenesisFilePath(s.dataDir))
	if err != nil {
		return errors.Wrap(err, "failed to load genesis file")
	}
	s.balances = genesis.Balances
	blocks, err := s.blockStore.Read(AfterNone, math.MaxUint64)
	if err != nil {
		return errors.Wrap(err, "failed to load blocks from block store")
	}
	for _, block := range blocks {
		if err := applyBlock(s, &block); err != nil {
			return errors.Wrap(err, "failed to apply block")
		}
	}
	return nil
}

func (s *State) GetBlocksAfter(after string) ([]Block, error) {
	return s.blockStore.Read(after, math.MaxUint64)
}

// TODO
func (s *State) AddBlock(block Block) error {
	return nil
}

func applyTx(s *State, tx Tx) error {
	if tx.IsReward() {
		s.balances[tx.To] += tx.Value
		return nil
	}
	if s.balances[tx.From] < tx.Value {
		return fmt.Errorf("insufficient balance")
	}
	s.balances[tx.From] -= tx.Value
	s.balances[tx.To] += tx.Value
	return nil
}

// TODO
func applyBlock(s *State, block *Block) error {
	return nil
}

func (s *State) LatestBlockHash() Hash {
	return s.lastBlockHash
}

func (s *State) LatestBlock() Block {
	return s.lastBlock
}

func (s *State) LatestBlockNumber() uint64 {
	return s.lastBlock.Header.Number
}
