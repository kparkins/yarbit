package database

import (
	"crypto/sha256"
	"encoding/json"
	"time"
)

type Tx struct {
	From  Account   `json:"from"`
	To    Account   `json:"to"`
	Value uint      `json:"value"`
	Data  string    `json:"data"`
	Time  time.Time `json:"time"`
}

func NewTx(from, to Account, value uint, data string) Tx {
	return Tx{
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
	}
}

func (t Tx) IsReward() bool {
	return t.Data == "reward"
}

func (t Tx) Hash() (Hash, error) {
	var hash Hash
	data, err := json.Marshal(t)
	if err != nil {
		return hash, err
	}
	return sha256.Sum256(data), nil
}
