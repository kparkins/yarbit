package database

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

const MaxBlocksPerRead = 100

type State struct {
	Balances      map[Account]uint
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
		Balances:      make(map[Account]uint, 0),
		txMempool:     make([]Tx, 0),
		blockStore:    NewFileBlockStore(blockDbPath),
		lastBlockHash: Hash{},
		lastBlock:     NewBlock(Hash{}, 0, 0, make([]Tx, 0)),
	}
	return state
}

func (s *State) Load() error {
	if err := initDataDir(s.dataDir); err != nil {
		return err
	}
	genesis, err := LoadGenesis(getGenesisFilePath(s.dataDir))
	if err != nil {
		return err
	}
	s.Balances = genesis.Balances
	group := sync.WaitGroup{}
	group.Add(1)
	blocks := make(chan Block, 100)
	go func() {
		s.blockStore.Stream(AfterNone, blocks)
		group.Done()
	}()
	for b := range blocks {
		s.AddBlock(b)
	}
	group.Wait()
	return nil
}

func (s *State) GetBlocksAfter(key string) ([]Block, error) {
	return s.blockStore.Read(key, MaxBlocksPerRead)
}

func (s *State) AddTx(tx Tx) error {
	if err := s.applyTx(tx); err != nil {
		return err
	}
	s.txMempool = append(s.txMempool, tx)
	return nil
}

func (s *State) AddBlock(block Block) error {
	for _, tx := range block.Txs {
		if err := s.AddTx(tx); err != nil {
			return err
		}
	}
	return nil
}

func (s *State) Persist() (Hash, error) {
	block := NewBlock(
		s.lastBlockHash,
		s.lastBlock.Header.Number,
		uint64(time.Now().Unix()),
		s.txMempool,
	)
	// Hack to force the sequence number to 0 for the first block. This is all because the way we are
	// adding blocks and saving state is buggy in this version. Need to refactor this file.
	if !reflect.DeepEqual(Hash{}, s.lastBlockHash) {
		block.Header.Number++
	}

	//s.lastBlockHash = hash
	s.lastBlock = block
	s.txMempool = make([]Tx, 0)
	return s.lastBlockHash, nil
}

func (s *State) applyTx(tx Tx) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}
	if s.Balances[tx.From] < tx.Value {
		return fmt.Errorf("insufficient balance")
	}
	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value
	return nil
}

func (s *State) applyBlock(block Block) error {
	for _, tx := range block.Txs {
		if err := s.applyTx(tx); err != nil {
			return err
		}
	}
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
