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

type Node struct {
	dataDir string
	port    uint64

	state      *database.State
	knownPeers []PeerNode
}

func New(dataDir string, port uint64, bootstrap PeerNode) *Node {
	node := &Node{
		dataDir:    dataDir,
		port:       port,
		knownPeers: make([]PeerNode, 0),
	}
	if bootstrap.Ip != "" {
		node.knownPeers = append(node.knownPeers, bootstrap)
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
		balanceListHandler(writer, request, n.state)
	})
	http.HandleFunc("/tx/add", func(writer http.ResponseWriter, request *http.Request) {
		txAddHandler(writer, request, n.state)
	})
	http.HandleFunc("/node/status", func(writer http.ResponseWriter, request *http.Request) {
		nodeStatusHandler(writer, request, n)
	})
    http.HandleFunc("/node/peers", func(writer http.ResponseWriter, request *http.Request) {
        nodePeersHandler(writer, request)
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

func (n *Node) KnownPeers() []PeerNode {
	return n.knownPeers
}
