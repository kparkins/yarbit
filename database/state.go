package database

import (
	"fmt"
	"math"
	"reflect"

	"github.com/pkg/errors"
)

const BlockReward = 10
const MaxBlocksPerRead = 1000

type State struct {
	balances      map[Account]uint
	dataDir       string
	blockStore    BlockStore
	lastBlockHash Hash
	lastBlock     *Block
	hasGenesis    bool
}

func NewStateFromDisk(dataDir string) *State {
	blockDbPath := getBlockDatabaseFilePath(dataDir)
	state := &State{
		dataDir:       dataDir,
		balances:      make(map[Account]uint, 0),
		blockStore:    NewFileBlockStore(blockDbPath),
		lastBlockHash: Hash{},
		lastBlock:     NewBlock(Hash{}, 0, 0, make([]Tx, 0)),
		hasGenesis:    false,
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
	blocks, err := s.blockStore.Read(AfterGenesis, math.MaxUint64)
	if err != nil {
		return errors.Wrap(err, "failed to load blocks from block store")
	}
	for _, block := range blocks {
		if err := s.ApplyBlock(&block); err != nil {
			return errors.Wrap(err, "failed to apply block")
		}
	}
	if len(blocks) <= 0 {
		return nil
	}
	s.hasGenesis = true
	last := blocks[len(blocks)-1]
	hash, err := last.Hash()
	if err != nil {
		return err
	}
	s.lastBlock = &last
	s.lastBlockHash = hash
	return nil
}

func (s *State) GetBlocksAfter(after string) ([]Block, error) {
	return s.blockStore.Read(after, math.MaxUint64)
}

func (s *State) NextBlockNumber() uint64 {
	if !s.hasGenesis {
		return uint64(0)
	}
	return s.lastBlock.Header.Number + 1
}

func (s *State) AddBlock(block *Block) (Hash, error) {
	var hash Hash
	if block.Header.Number != s.NextBlockNumber() {
		return hash, fmt.Errorf("new block doesn't have the correct sequence number")
	}
	if !reflect.DeepEqual(block.Header.Parent, s.lastBlockHash) {
		return hash, fmt.Errorf("new block doesn't have the correct parent hash")
	}
	c := s.Clone()
	if err := c.ApplyBlock(block); err != nil {
		return hash, errors.Wrap(err, "failed to apply block")
	}
	hash, err := s.blockStore.Write(block)
	if err != nil {
		return hash, errors.Wrap(err, "could not persist new block to data store")
	}
	fmt.Printf("Saved new block to storage: \n")
	fmt.Printf("\t%s\n", hash.String())
	s.hasGenesis = true
	s.balances = c.balances
	s.lastBlockHash = hash
	s.lastBlock = block
	return hash, nil
}

func (s *State) Clone() *State {
	return &State{
		balances:      s.Balances(),
		dataDir:       s.dataDir,
		blockStore:    s.blockStore,
		lastBlock:     s.lastBlock.Clone(),
		lastBlockHash: s.lastBlockHash.Clone(),
		hasGenesis:    s.hasGenesis,
	}
}

func (s *State)ApplyBlock(block *Block) error {
	for _, tx := range block.Txs {
		if err := s.ApplyTx(tx); err != nil {
			return err
		}
	}
	s.balances[block.Header.Miner] += BlockReward
	return nil
}

func (s *State) ApplyTx(tx Tx) error {
	if tx.IsReward() {
		s.balances[tx.To] += tx.Value
		return nil
	}
	txHash, _ := tx.Hash()
	if s.balances[tx.From] < tx.Value {
		return fmt.Errorf("TX: %s insufficient balance", txHash)
	}
	s.balances[tx.From] -= tx.Value
	s.balances[tx.To] += tx.Value
	return nil
}

func (s *State) LatestBlockHash() Hash {
	return s.lastBlockHash
}

func (s *State) LatestBlock() *Block {
	return s.lastBlock.Clone()
}

func (s *State) LatestBlockNumber() uint64 {
	return s.lastBlock.Header.Number
}
