package database

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

const GenesisJson = `{
  "genesis_time": "2021-03-17T02:59:45.008883+00:00",
  "chain_id": "the-yarbit-ledger",
  "balances": {
    "andrej": 1000000
  }
}
`

type Genesis struct {
	GenesisTime time.Time        `json:"genesis_time"`
	ChainId     string           `json:"chain_id"`
	Balances    map[Account]uint `json:"balances"`
}

func LoadGenesis(path string) (*Genesis, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	genesis := &Genesis{}
	if err = json.Unmarshal(content, genesis); err != nil {
		return nil, err
	}
	return genesis, nil
}

func writeGenesisToDisk(path string) error {
	if err := ioutil.WriteFile(path, []byte(GenesisJson), 0644); err != nil {
		return err
	}
	return os.Chown(path, os.Getuid(), os.Getgid())
}
