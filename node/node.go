package node

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kparkins/yarbit/database"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type Node struct {
	config     Config
	router     *mux.Router
	state      *database.State
	knownPeers map[string]PeerNode
}

func New(dataDir string, port uint64, bootstrap PeerNode) *Node {
	node := &Node{
		config: Config{
			DataDir: dataDir,
			Port:    port,
		},
		router:     mux.NewRouter(),
		knownPeers: make(map[string]PeerNode, 0),
	}
	if bootstrap.Ip != "" {
		node.knownPeers[bootstrap.SocketAddress()] = bootstrap
	}
	return node
}

func (n *Node) routes() {
	n.router.HandleFunc("/tx/add", n.handleAddTx())
	n.router.HandleFunc("/node/peer", n.handleAddPeer())
	n.router.HandleFunc("/node/sync", n.handleNodeSync())
	n.router.HandleFunc("/node/status", n.handleNodeStatus())
	n.router.HandleFunc("/balances/list", n.handleListBalances())
}

func (n *Node) Run() error {
	var err error
	fmt.Print("Loading state from disk...")
	n.state, err = database.NewStateFromDisk(n.config.DataDir)
	if err != nil {
		return errors.Wrap(err, "Failed to load state from disk.")
	}
	fmt.Print("Complete.\n")
	defer n.state.Close()
	n.routes()
	fmt.Printf("Listening on port: %d\n", n.config.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", n.config.Port), nil)
}

func (n *Node) handleListBalances() http.HandlerFunc {
	type BalancesListResponse struct {
		Hash     database.Hash             `json:"block_hash"`
		Balances map[database.Account]uint `json:"balances"`
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		writeResponse(writer, BalancesListResponse{
			Hash:     n.state.LatestBlockHash(),
			Balances: n.state.Balances,
		})
	}
}

func (n *Node) handleAddTx() http.HandlerFunc {
	type TxAddRequest struct {
		From  string `json:"from"`
		To    string `json:"to"`
		Value uint   `json:"value"`
		Data  string `json:"data"`
	}
	type TxAddResponse struct {
		Hash database.Hash `json:"block_hash"`
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		content, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writeErrorResponse(writer, err, http.StatusInternalServerError)
			return
		}
		defer request.Body.Close()
		txRequest := TxAddRequest{}
		if err := json.Unmarshal(content, &txRequest); err != nil {
			writeErrorResponse(writer, err, http.StatusBadRequest)
			return
		}
		tx := database.NewTx(
			database.NewAccount(txRequest.From),
			database.NewAccount(txRequest.To),
			txRequest.Value,
			txRequest.Data,
		)
		if err := n.state.AddTx(tx); err != nil {
			writeErrorResponse(writer, err, http.StatusInternalServerError)
			return
		}
		hash, err := n.state.Persist()
		if err != nil {
			writeErrorResponse(writer, err, http.StatusInternalServerError)
			return
		}
		writeResponse(writer, TxAddResponse{Hash: hash})
	}
}

func (n *Node) handleNodeStatus() http.HandlerFunc {
	type NodeStatusResponse struct {
		Hash       database.Hash       `json:"block_hash"`
		Number     uint64              `json:"block_number"`
		KnownPeers map[string]PeerNode `json:"known_peers"`
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		writeResponse(writer, NodeStatusResponse{
			Hash:       n.state.LatestBlockHash(),
			Number:     n.state.LatestBlockNumber(),
			KnownPeers: n.knownPeers,
		})
	}
}

func (n *Node) handleNodeSync() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

	}
}

func (n *Node) handleAddPeer() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

	}
}
