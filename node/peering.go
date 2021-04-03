package node

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type PeerService struct {
	peers map[string]PeerNode
	lock  *sync.RWMutex
	host  PeerNode
}

func NewPeerService(host PeerNode) *PeerService {
	return &PeerService{
		peers: make(map[string]PeerNode),
		lock:  &sync.RWMutex{},
		host:  host,
	}
}

func (p *PeerService) AddPeer(peer PeerNode) bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	if !peer.IsActive {
		return false
	}
	address := peer.SocketAddress()
	nodeAddress := p.host.SocketAddress()
	if _, ok := p.peers[address]; ok || address == nodeAddress {
		return false
	}
	p.peers[address] = peer
	fmt.Printf("added new peer %s\n", address)
	return true
}

func (p *PeerService) AddPeers(peers map[string]PeerNode) {
	for _, v := range peers {
		p.AddPeer(v)
	}
}

func (p *PeerService) RemovePeer(peer PeerNode) {
	p.lock.Lock()
	defer p.lock.Unlock()
	address := peer.SocketAddress()
	delete(p.peers, address)
	fmt.Fprintf(os.Stderr, "removed %s from known peers\n", address)
}

func (p *PeerService) Peers() map[string]PeerNode {
	p.lock.RLock()
	defer p.lock.RUnlock()
	peers := make(map[string]PeerNode, len(p.peers))
	for k, v := range p.peers {
		peers[k] = v
	}
	return peers
}

func (p *PeerService) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	c, cancel := context.WithCancel(ctx)
	for {
		select {
		case <-ticker.C:
			p.mingle(c)
		case <-ctx.Done():
			cancel()
			return
		}
	}
}

func (p *PeerService) mingle(ctx context.Context) {
	knownPeers := p.Peers()
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	nodeAddress := p.host.SocketAddress()
	for _, peer := range knownPeers {
		peerAddress := peer.SocketAddress()
		status, err := fetchPeerStatus(ctx, client, peerAddress)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v while checking status of %s\n", err, peerAddress)
			p.RemovePeer(peer)
			continue
		}
		status.KnownPeers = FilterPeers(status.KnownPeers, func(s string) bool {
			return s != nodeAddress
		})
		p.AddPeers(status.KnownPeers)
		// TODO
		//p.AddPendingTxs(status.PendingTxs)
		if err := joinPeers(ctx, client, peer, p.host); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		latestHash := n.LatestBlockHash()
		if (status.Number == 0 && !latestHash.IsEmpty()) || status.Number < n.LatestBlockNumber() {
			continue
		}
		blocks, err := fetchBlocks(ctx, client, peerAddress, n.LatestBlockHash())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		for i := range blocks {
			n.newBlockChan <- &blocks[i]
		}

	}

}

func fetchPeerStatus(ctx context.Context, client *http.Client, address string) (StatusResponse, error) {
	var status StatusResponse
	url := fmt.Sprintf("%s://%s%s", "http", address, ApiRouteStatus)
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return status, errors.Wrap(err, "while creating request")
	}
	response, err := client.Do(request)
	if err != nil {
		return status, errors.Wrap(err, fmt.Sprintf("error fetching peers from %s", address))
	}
	statusResponse := StatusResponse{}
	if err := readJsonResponse(response, &statusResponse); err != nil {
		return status, err
	}
	return statusResponse, nil
}

func joinPeers(ctx context.Context, client *http.Client, peer, host PeerNode) error {
	url := fmt.Sprintf("%s://%s%s", "http", peer.SocketAddress(), ApiRouteAddPeer)
	body, err := json.Marshal(host)
	if err != nil {
		return fmt.Errorf("error marshaling add peer request body")
	}
	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "while creating request")
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error joining peers from %s", peer.SocketAddress()))
	}
	_ = response.Body.Close()
	return nil
}
