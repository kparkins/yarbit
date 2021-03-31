package node

import "github.com/kparkins/yarbit/database"

const (
	ApiRouteAddPeer      = "/node/peer"
	ApiRouteSync         = "/node/sync"
	ApiRouteAddTx        = "/tx/add"
	ApiRouteStatus       = "/node/status"
	ApiRouteListBalances = "/balances/list"

	ApiQueryParamAfter = "after"
)

type StatusResponse struct {
	Hash       database.Hash       `json:"block_hash"`
	Number     uint64              `json:"block_number"`
	KnownPeers map[string]PeerNode `json:"known_peers"`
	PendingTxs []database.Tx       `json:"pending_txs"`
}

type AddPeerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type SyncResult struct {
	Blocks []database.Block `json:"blocks"`
}

type TxAddRequest struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}
