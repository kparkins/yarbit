package node

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kparkins/yarbit/database"
	"github.com/pkg/errors"
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
	if bootstrap.IpAddress != "" {
		node.knownPeers[bootstrap.SocketAddress()] = bootstrap
	}
	return node
}

func (n *Node) routes() {
	n.router.HandleFunc("/tx/add", n.handleAddTx()).Methods("POST")
	n.router.HandleFunc("/node/peer", n.handleAddPeer()).Methods("POST")
	n.router.HandleFunc("/node/sync", n.handleNodeSync()).Methods("GET")
	n.router.HandleFunc("/node/status", n.handleNodeStatus()).Methods("GET")
	n.router.HandleFunc("/balances/list", n.handleListBalances()).Methods("GET")
}

func (n *Node) Run() error {
	fmt.Print("Loading state from disk...")
	n.state = database.NewStateFromDisk(n.config.DataDir)
	if err := n.state.Load(); err != nil {
		return errors.Wrap(err, "Failed to load state from disk.")
	}
	fmt.Print("Complete.\n")
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
			Balances: n.state.Balances(),
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
		var txRequest TxAddRequest
		err := readRequestJson(request, &txRequest)
		if err != nil {
			writeErrorResponse(writer, err, http.StatusBadRequest)
			return
		}
		defer request.Body.Close()
		_ = database.NewTx(
			database.NewAccount(txRequest.From),
			database.NewAccount(txRequest.To),
			txRequest.Value,
			txRequest.Data,
		)
		//TODO
		/*
		if err := n.state.AddTx(tx); err != nil {
			writeErrorResponse(writer, err, http.StatusInternalServerError)
			return
		}
		hash, err := n.state.Persist()*/
		if err != nil {
			writeErrorResponse(writer, err, http.StatusInternalServerError)
			return
		}
		//writeResponse(writer, TxAddResponse{Hash: hash})
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
		after := request.URL.Query().Get("after")
		fmt.Println(after)
	}
}

func (n *Node) handleAddPeer() http.HandlerFunc {
	type AddPeerResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		var peer PeerNode
		if err := readRequestJson(request, &peer); err != nil {
			writeErrorResponse(writer, err, http.StatusBadRequest)
			return
		}
		peer.IsActive = true
		if _, ok := n.knownPeers[peer.SocketAddress()]; !ok {
			n.knownPeers[peer.SocketAddress()] = peer
		}
		writeResponse(writer, AddPeerResponse{
			Success: true,
			Message: fmt.Sprintf("Added %s to known peers", peer.SocketAddress()),
		})
	}
}
