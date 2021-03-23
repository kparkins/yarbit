package node

import (
	"encoding/json"
	"github.com/kparkins/yarbit/database"
	"io/ioutil"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type BalancesListResponse struct {
	Hash     database.Hash             `json:"block_hash"`
	Balances map[database.Account]uint `json:"balances"`
}

type TxAddRequest struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

type TxAddResponse struct {
	Hash database.Hash `json:"block_hash"`
}

type NodeStatusResponse struct {
	Hash       database.Hash `json:"block_hash"`
	Number     uint64        `json:"block_number"`
	KnownPeers []PeerNode    `json:"known_peers"`
}

func balanceListHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	writeResponse(w, BalancesListResponse{Hash: state.LatestBlockHash(), Balances: state.Balances})
}

func txAddHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	txRequest := TxAddRequest{}
	if err := json.Unmarshal(content, &txRequest); err != nil {
		writeErrorResponse(w, err, http.StatusBadRequest)
		return
	}
	tx := database.NewTx(
		database.NewAccount(txRequest.From),
		database.NewAccount(txRequest.To),
		txRequest.Value,
		txRequest.Data,
	)
	if err := state.AddTx(tx); err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	hash, err := state.Persist()
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	writeResponse(w, TxAddResponse{Hash: hash})
}

func nodeStatusHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	response := NodeStatusResponse{
		Hash:   node.LatestBlockHash(),
		Number: node.LatestBlockNumber(),
		KnownPeers: node.KnownPeers(),
	}
	writeResponse(w, response)
}

func nodePeersHandler(w http.ResponseWriter, r *http.Request) {
}
