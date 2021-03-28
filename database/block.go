package database

import (
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

func (b *Block) Hash() (Hash, error) {
	encoded, err := json.Marshal(*b)
	if err != nil {
		return Hash{}, err
	}
	return sha256.Sum256(encoded), nil
}

func (b *Block) Clone() *Block {
	txs := make([]Tx, len(b.Txs))
	copy(txs, b.Txs)
	return &Block{
		Header: b.Header.Clone(),
		Txs:    txs,
	}
}

func IsBlockHashValid(hash Hash) bool {
	return hash.String()[:6] == "000000"
}
