package node

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/kparkins/yarbit/database"
	"github.com/pkg/errors"
)

type Node struct {
	config        Config
	lock          *sync.RWMutex
	router        *mux.Router
	state         *database.State
	pendingTxs    map[database.Hash]database.Tx
	completedTxs  map[database.Hash]database.Tx // TODO need to expire or write to disk periodically
	peering       *PeerService
	miner         *Miner
	server        *http.Server
	newBlockChan  chan *database.Block
	miningAccount database.Account
}

func New(config Config) *Node {
	blockChan := make(chan *database.Block)
	node := &Node{
		config:       config,
		lock:         &sync.RWMutex{},
		router:       mux.NewRouter(),
		pendingTxs:   make(map[database.Hash]database.Tx),
		completedTxs: make(map[database.Hash]database.Tx),
		peering:      NewPeerService(NewPeerNode(config.IpAddress, config.Port)),
		miner:        NewMiner(blockChan),
		server:       &http.Server{},
		newBlockChan: blockChan,
	}
	if config.Bootstrap.IpAddress != "" {
		node.peering.AddPeer(config.Bootstrap)
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
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		fmt.Printf("Listening at: %s:%d\n", n.config.IpAddress, n.config.Port)
		n.server.ListenAndServe()
	}()
	go n.peering.Start(ctx)
	go n.miner.Start(ctx)
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
	type TxAddResponse struct {
		Hash database.Hash `json:"tx_hash"`
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
		hash, err := n.AddPendingTx(tx)
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
			KnownPeers: n.peering.Peers(),
			PendingTxs: n.PendingTxs(),
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
		n.peering.AddPeer(peer)
		writeJsonResponse(writer, AddPeerResponse{
			Success: true,
			Message: fmt.Sprintf("added %s to known peers", peer.SocketAddress()),
		})
	}
}

func (n *Node) RemoveTxs(txs []database.Tx) {
	for _, tx := range txs {
		hash, err := tx.Hash()
		if err != nil {
			fmt.Println(err)
			continue
		}
		delete(n.pendingTxs, hash)
	}
}

func (n *Node) LatestBlockNumber() uint64 {
	return n.state.LatestBlockNumber()
}

func (n *Node) Balances() map[database.Account]uint {
	return n.state.Balances()
}

func (n *Node) LatestBlockHash() database.Hash {
	return n.state.LatestBlockHash()
}

func (n *Node) Protocol() string {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.config.Protocol
}

func (n *Node) AddBlock(block *database.Block) (database.Hash, error) {
	return n.state.AddBlock(block)
}

func (n *Node) CompleteTxs(txs []database.Tx) error {
	n.lock.Lock()
	defer n.lock.Unlock()
	for _, tx := range txs {
		hash, err := tx.Hash()
		if err != nil {
			return err
		}
		n.completedTxs[hash] = tx
		delete(n.pendingTxs, hash)
	}
	return nil
}

func (n *Node) AddPendingTx(tx database.Tx) (database.Hash, error) {
	n.lock.Lock()
	defer n.lock.Unlock()
	var hash database.Hash
	hash, err := tx.Hash()
	if err != nil {
		fmt.Printf("error hashing new tx %v\n", tx)
		return hash, err
	}
	if _, ok := n.completedTxs[hash]; ok {
		return hash, nil
	}
	n.pendingTxs[hash] = tx
	return hash, nil
}

func (n *Node) AddPendingTxs(txs []database.Tx) error {
	for _, tx := range txs {
		n.AddPendingTx(tx)
	}
	return nil
}

func (n *Node) PendingTxs() []database.Tx {
	n.lock.RLock()
	defer n.lock.RUnlock()
	txs := make([]database.Tx, 0, len(n.pendingTxs))
	for _, v := range n.pendingTxs {
		txs = append(txs, v)
	}
	return txs
}

func (n *Node) GetBlocksAfter(after string) ([]database.Block, error) {
	return n.state.GetBlocksAfter(after)
}

func (n *Node) createPendingBlock() *database.Block {
	n.lock.RLock()
	defer n.lock.RUnlock()
	txs := make([]database.Tx, 0, len(n.pendingTxs))
	for _, tx := range n.pendingTxs {
		txs = append(txs, tx)
	}
	return &database.Block{
		Header: database.BlockHeader{
			Parent: n.state.LatestBlockHash(),
			Number: n.state.NextBlockNumber(),
			Nonce:  0,
			Time:   uint64(time.Now().Unix()),
			Miner:  n.config.MinerAccount,
		},
		Txs: txs,
	}
}
