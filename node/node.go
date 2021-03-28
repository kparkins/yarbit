package node

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kparkins/yarbit/database"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Node struct {
	config     Config
	protocol   string
	lock       *sync.RWMutex
	router     *mux.Router
	state      *database.State
	txPool     []database.Tx
	knownPeers map[string]PeerNode
	server     *http.Server
}

func New(dataDir string, ip string, port uint64, bootstrap PeerNode) *Node {
	node := &Node{
		config: Config{
			DataDir:   dataDir,
			IpAddress: ip,
			Port:      port,
		},
		protocol:   "http",
		lock:       &sync.RWMutex{},
		router:     mux.NewRouter(),
		txPool:     make([]database.Tx, 0),
		knownPeers: make(map[string]PeerNode, 0),
		server:     &http.Server{},
	}
	if bootstrap.IpAddress != "" {
		node.knownPeers[bootstrap.SocketAddress()] = bootstrap
	}
	node.routes()
	node.server.Addr = fmt.Sprintf(":%d", node.config.Port)
	node.server.Handler = node.router
	return node
}

func (n *Node) routes() {
	n.router.HandleFunc(ApiRouteAddTx, n.handleAddTx()).Methods("POST")
	n.router.HandleFunc(ApiRouteAddPeer, n.handleAddPeer()).Methods("POST")
	n.router.HandleFunc(ApiRouteSync, n.handleNodeSync()).Methods("GET")
	n.router.HandleFunc(ApiRouteStatus, n.handleNodeStatus()).Methods("GET")
	n.router.HandleFunc(ApiRouteListBalances, n.handleListBalances()).Methods("GET")
}

func (n *Node) Run() error {
	fmt.Print("Loading state from disk...")
	n.state = database.NewStateFromDisk(n.config.DataDir)
	if err := n.state.Load(); err != nil {
		return errors.Wrap(err, "Failed to load state from disk.")
	}
	fmt.Print("Complete.\n")
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		fmt.Printf("Listening at: %s:%d\n", n.config.IpAddress, n.config.Port)
		n.server.ListenAndServe()
	}()
	go n.sync(ctx)
	go n.mine(ctx)
	<-quit
	cancel()
	n.server.Shutdown(ctx)
	return nil
}

func (n *Node) handleListBalances() http.HandlerFunc {
	type BalancesListResponse struct {
		Hash     database.Hash             `json:"block_hash"`
		Balances map[database.Account]uint `json:"balances"`
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		writeJsonResponse(writer, BalancesListResponse{
			Hash:     n.LatestBlockHash(),
			Balances: n.Balances(),
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
		err := readJsonRequest(request, &txRequest)
		if err != nil {
			writeJsonErrorResponse(writer, err, http.StatusBadRequest)
			return
		}
		defer request.Body.Close()
		tx := database.NewTx(
			database.NewAccount(txRequest.From),
			database.NewAccount(txRequest.To),
			txRequest.Value,
			txRequest.Data,
		)
		block := database.NewBlock(
			n.LatestBlockHash(),
			n.LatestBlockNumber()+1,
			uint64(time.Now().Unix()),
			[]database.Tx{tx},
		)
		hash, err := n.AddBlock(block)
		if err != nil {
			writeJsonErrorResponse(writer, err, http.StatusInternalServerError)
			return
		}
		writeJsonResponse(writer, TxAddResponse{Hash: hash})
	}
}

func (n *Node) handleNodeStatus() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writeJsonResponse(writer, StatusResponse{
			Hash:       n.LatestBlockHash(),
			Number:     n.LatestBlockNumber(),
			KnownPeers: n.Peers(),
		})
	}
}

func (n *Node) handleNodeSync() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		after := request.URL.Query().Get(ApiQueryParamAfter)
		blocks, err := n.GetBlocksAfter(after)
		if err != nil {
			writeJsonErrorResponse(writer, err, http.StatusInternalServerError)
			return
		}
		writeJsonResponse(writer, SyncResult{Blocks: blocks})
	}
}

func (n *Node) handleAddPeer() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var peer PeerNode
		if err := readJsonRequest(request, &peer); err != nil {
			writeJsonErrorResponse(writer, err, http.StatusBadRequest)
			return
		}
		n.AddPeer(peer)
		writeJsonResponse(writer, AddPeerResponse{
			Success: true,
			Message: fmt.Sprintf("added %s to known peers", peer.SocketAddress()),
		})
	}
}

func (n *Node) sync(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			syncWithPeers(ctx, n)
			break
		case <-ctx.Done():
			return
		}
	}
}

func (n *Node) mine(ctx context.Context, ) {
	c, cancelMiner := context.WithCancel(ctx)
	for {
		select {
		case <-ctx.Done():
			cancelMiner()
			return
		default:
			mine(c, nil)
		}
	}
}

func mine(ctx context.Context, pending *database.Block) {

}

func (n *Node) LatestBlockNumber() uint64 {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.state.LatestBlockNumber()
}

func (n *Node) Balances() map[database.Account]uint {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.state.Balances()
}

func (n *Node) LatestBlockHash() database.Hash {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.state.LatestBlockHash()
}

func (n *Node) Protocol() string {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.protocol
}

func (n *Node) AddPeer(peer PeerNode) bool {
	n.lock.Lock()
	defer n.lock.Unlock()
	if !peer.IsActive {
		return false
	}
	address := peer.SocketAddress()
	nodeAddress := fmt.Sprintf("%s:%d", n.config.IpAddress, n.config.Port)
	if _, ok := n.knownPeers[address]; ok || address == nodeAddress {
		return false
	}
	n.knownPeers[address] = peer
	fmt.Printf("added new peer %s\n", address)
	return true
}

func (n *Node) AddPeers(peers map[string]PeerNode) {
	for _, v := range peers {
		n.AddPeer(v)
	}
}

func (n *Node) RemovePeer(peer PeerNode) {
	n.lock.Lock()
	defer n.lock.Unlock()
	address := peer.SocketAddress()
	delete(n.knownPeers, address)
	fmt.Fprintf(os.Stderr, "removed %s from known peers\n", address)
}

func (n *Node) Peers() map[string]PeerNode {
	n.lock.RLock()
	defer n.lock.RUnlock()
	peers := make(map[string]PeerNode, len(n.knownPeers))
	for k, v := range n.knownPeers {
		peers[k] = v
	}
	return peers
}

func (n *Node) AddBlock(block *database.Block) (database.Hash, error) {
	n.lock.Lock()
	defer n.lock.Unlock()
	return n.state.AddBlock(block)
}

func (n *Node) GetBlocksAfter(after string) ([]database.Block, error) {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.state.GetBlocksAfter(after)
}
