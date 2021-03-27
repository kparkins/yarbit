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
	Parent Hash   `json:"parent"`
	Number uint64 `json:"number"`
	None   int32  `json:"nonce"`
	Time   uint64 `json:"time"`
}

func (h BlockHeader) Clone() BlockHeader {
	return BlockHeader{
		Parent: h.Parent.Clone(),
		Number: h.Number,
		Time:   h.Time,
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
