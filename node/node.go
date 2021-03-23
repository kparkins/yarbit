package node

import (
	"fmt"
	"github.com/kparkins/yarbit/database"
	"net/http"
)

type PeerNode struct {
	Ip          string `json:"ip"`
	Port        uint64 `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`
	IsActive    bool   `json:"is_active"`
}

func (p PeerNode) SocketAddress() string {
	return fmt.Sprintf("%s:%d", p.Ip, p.Port)
}

type Node struct {
	dataDir string
	port    uint64

	state      *database.State
	knownPeers map[string]PeerNode
}

func New(dataDir string, port uint64, bootstrap PeerNode) *Node {
	node := &Node{
		dataDir:    dataDir,
		port:       port,
		knownPeers: make(map[string]PeerNode, 0),
	}
	if bootstrap.Ip != "" {
		node.knownPeers[bootstrap.SocketAddress()] = bootstrap
	}
	return node
}

func (n *Node) Run() error {
	fmt.Print("Loading state from disk...")
	var err error
	n.state, err = database.NewStateFromDisk(n.dataDir)
	if err != nil {
		fmt.Print("Failed.\n")
		return err
	}
	fmt.Print("Complete.\n")
	defer n.state.Close()

	http.HandleFunc("/balances/list", func(writer http.ResponseWriter, request *http.Request) {
		handleListBalances(writer, request, n.state)
	})
	http.HandleFunc("/tx/add", func(writer http.ResponseWriter, request *http.Request) {
		handleAddTx(writer, request, n.state)
	})
	http.HandleFunc("/node/status", func(writer http.ResponseWriter, request *http.Request) {
		handleNodeStatus(writer, request, n)
	})
	http.HandleFunc("/node/sync", func(writer http.ResponseWriter, request *http.Request) {
		handleNodeSync(writer, request)
	})
	http.HandleFunc("/node/peer", func(writer http.ResponseWriter, request *http.Request) {
		handleAddPeer(writer, request)
	})

	fmt.Printf("Listening on port: %d\n", n.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", n.port), nil)
}

func (n *Node) LatestBlockHash() database.Hash {
	return n.state.LatestBlockHash()
}

func (n *Node) LatestBlockNumber() uint64 {
	return n.state.LatestBlockNumber()
}

func (n *Node) KnownPeers() map[string]PeerNode {
	return n.knownPeers
}
