package database

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
)

type BlockFileEntry struct {
	Hash  Hash   `json:"hash"`
	Block *Block `json:"block"`
}

type BlockHeader struct {
	Parent Hash    `json:"parent"`
	Number uint64  `json:"number"`
	Nonce  int32   `json:"nonce"`
	Time   uint64  `json:"time"`
	Miner  Account `json:"miner"`
}

func (h BlockHeader) Clone() BlockHeader {
	return BlockHeader{
		Parent: h.Parent.Clone(),
		Number: h.Number,
		Nonce:  h.Nonce,
		Time:   h.Time,
		Miner:  h.Miner,
	}
}

type Block struct {
	Header BlockHeader `json:"header"`
	Txs    []Tx        `json:"payload"`
}

func NewBlock(parent Hash, number, time uint64, txs []Tx) *Block {
	return &Block{
		Header: BlockHeader{
			Parent: parent,
			Number: number,
			Nonce:  0,
			Time:   time,
		},
		Txs: txs,
	}
}

func (b *Block) DebugString() string {
	txs := make([]Hash, 0)
	for _, tx := range b.Txs {
		if hash, err := tx.Hash(); err != nil {
			txs = append(txs, hash)
		}
	}
	out := struct {
		Header BlockHeader `json:"header"`
		Hashes    []Hash        `json:"payload"`
	} {
		Header: b.Header,
		Hashes: txs,
	}
	if json, err := json.Marshal(&out); err == nil {
		return string(json) 
	}
	return "" 
}

func (b *Block) Hash() (Hash, error) {
	encoded, err := json.Marshal(*b)
	if err != nil {
		return Hash{}, err
	}
	return sha256.Sum256(encoded), nil
}

func (b *Block) Clone() *Block {
	txs := make([]Tx, 0, len(b.Txs))
	copy(txs, b.Txs)
	return &Block{
		Header: b.Header.Clone(),
		Txs:    txs,
	}
}

var MiningDifficulty = 3
var MiningDifficultyBytes []byte

func init() {
	for i := 0; i < MiningDifficulty; i++ {
		MiningDifficultyBytes = append(MiningDifficultyBytes, byte(0))
	}
}

func IsBlockHashValid(hash Hash) bool {
	return bytes.Equal(hash[:MiningDifficulty], MiningDifficultyBytes)
}
