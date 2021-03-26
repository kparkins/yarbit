package node

import (
	"context"
	"fmt"
	"github.com/kparkins/yarbit/database"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"time"
)

func runSync(ctx context.Context, n *Node) {
	ticker := time.NewTicker(45 * time.Second)
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

func syncWithPeers(ctx context.Context, n *Node) {
	knownPeers := n.Peers()
	for _, peer := range knownPeers {
		status, err := fetchPeerStatus(ctx, peer)
		if err != nil {
			fmt.Fprintln(os.Stderr, "%v while checking status of %s", err, peer.SocketAddress())
			n.RemovePeer(peer)
			continue
		}
		n.AddPeers(status.KnownPeers)
		joinPeers(ctx, peer, n.config.IpAddress, n.config.Port)
		if status.Number <= n.LatestBlockNumber() {
			continue
		}
		fetchBlocks(ctx, peer, n.LatestBlockHash())
	}

}

func fetchPeerStatus(ctx context.Context, peer PeerNode) (StatusResponse, error) {
	client := &http.Client{}
	var status StatusResponse
	address := peer.SocketAddress()
	if !peer.IsActive {
		return status, fmt.Errorf("%s not active", address)
	}
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	url := fmt.Sprintf("%s://%s/node/status", "http", address)
	req, err := http.NewRequestWithContext(timeout, "GET", url, nil)
	if err != nil {
		return status, errors.Wrap(err, "while creating request")
	}
	response, err := client.Do(req)
	if err != nil {
		return status, errors.Wrap(err, fmt.Sprintf("error fetching peers from %s", address))
	}
	statusResponse := StatusResponse{}
	if err := readResponseJson(response, &statusResponse); err != nil {
		return status, err
	}
	return statusResponse, nil
}

func joinPeers(ctx context.Context, peer PeerNode, ip string, port uint64) error {
	client := &http.Client{}
	address := peer.SocketAddress()
	if !peer.IsActive {
		return fmt.Errorf("%s not active", address)
	}
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	url := fmt.Sprintf("%s://%s/node/status", "http", address)
	req, err := http.NewRequestWithContext(timeout, "GET", url, nil)
	if err != nil {
		return errors.Wrap(err, "while creating request")
	}
	response, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error fetching peers from %s", address))
	}
	statusResponse := StatusResponse{}
	if err := readResponseJson(response, &statusResponse); err != nil {
		return err
	}
	return nil
}

func fetchBlocks(ctx context.Context, peer PeerNode, after database.Hash) {
}
