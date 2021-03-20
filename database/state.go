package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type State struct {
	Balances      map[Account]uint
	txMempool     []Tx
	blockDb       *os.File
	lastBlockHash Hash
}

func NewStateFromDisk(dataDir string) (*State, error) {
	if err := initDataDir(dataDir); err != nil {
		return nil, err
	}
	genesis, err := LoadGenesis(getGenesisFilePath(dataDir))
	if err != nil {
		return nil, err
	}
	blockDbPath := getBlockDatabaseFilePath(dataDir)
	blockDb, err := os.OpenFile(blockDbPath, os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}
	state := &State{
		Balances:      genesis.Balances,
		txMempool:     make([]Tx, 0),
		blockDb:       blockDb,
		lastBlockHash: Hash{},
	}
	scanner := bufio.NewScanner(blockDb)
	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return nil, err
		}
		var blockFileObject BlockFs
		if err = json.Unmarshal(scanner.Bytes(), &blockFileObject); err != nil {
			return nil, err
		}
		if err = state.applyBlocK(blockFileObject.Value); err != nil {
			return nil, err
		}
		state.lastBlockHash = blockFileObject.Key
	}
	return state, nil
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
	block := NewBlock(s.lastBlockHash, uint64(time.Now().Unix()), s.txMempool)
	hash, err := block.Hash()
	if err != nil {
		return Hash{}, err
	}
	blockFile := BlockFs{
		Key: hash,
		Value: block,
	}
	blockFileJson, err := json.Marshal(blockFile)
	if err != nil {
		return Hash{}, err
	}
	fmt.Printf("Persisting new block to disk: \n")
	fmt.Printf("\t%x\n", hash)
	if _, err := s.blockDb.Write(append(blockFileJson, '\n')); err != nil {
		return Hash{}, err
	}
	s.lastBlockHash = hash
	s.txMempool = make([]Tx, 0)
	return s.lastBlockHash, nil
}

func (s *State) Close() error {
	return s.blockDb.Close()
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

func (s *State) applyBlocK(block Block) error {
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
