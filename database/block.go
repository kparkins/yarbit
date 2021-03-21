package database

import (
	"crypto/sha256"
	"encoding/json"
)

type BlockFs struct {
	Key   Hash  `json:"hash"`
	Value Block `json:"block"`
}

type BlockHeader struct {
	Parent Hash   `json:"parent"`
	Time   uint64 `json:"time"`
}

type Block struct {
	Header BlockHeader `json:"header"`
	Txs    []Tx        `json:"payload"`
}

func NewBlock(parent Hash, time uint64, txs []Tx) Block {
	return Block{
		Header: BlockHeader{
			Parent: parent,
			Time:   time,
		},
		Txs: txs,
	}
}

func (b *Block) Hash() (Hash, error) {
	encoded, err := json.Marshal(b)
	if err != nil {
		return Hash{}, err
	}
	return sha256.Sum256(encoded), nil
}

