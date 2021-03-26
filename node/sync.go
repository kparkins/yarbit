package node

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kparkins/yarbit/database"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"time"
)

func runSync(ctx context.Context, n *Node) {
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

func syncWithPeers(ctx context.Context, n *Node) {
	knownPeers := n.Peers()
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	for _, peer := range knownPeers {
		status, err := fetchPeerStatus(client, peer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v while checking status of %s\n", err, peer.SocketAddress())
			n.RemovePeer(peer)
			continue
		}
		n.AddPeers(status.KnownPeers)
		if err := joinPeers(client, peer, n.config.IpAddress, n.config.Port); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		if status.Number <= n.LatestBlockNumber() {
			continue
		}
		blocks, err := fetchBlocks(client, peer, n.LatestBlockHash())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		for i := range blocks {
			if _, err := n.AddBlock(&blocks[i]); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}

}

func fetchPeerStatus(client *http.Client, peer PeerNode) (StatusResponse, error) {
	var status StatusResponse
	address := peer.SocketAddress()
	if !peer.IsActive {
		return status, fmt.Errorf("%s not active", address)
	}
	url := fmt.Sprintf("%s://%s%s", "http", address, ApiRouteStatus)
	req, err := http.NewRequest("GET", url, nil)
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

func joinPeers(client *http.Client, peer PeerNode, ip string, port uint64) error {
	address := peer.SocketAddress()
	if !peer.IsActive {
		return fmt.Errorf("%s not active", address)
	}
	url := fmt.Sprintf("%s://%s%s", "http", address, ApiRouteAddPeer)
	body, err := json.Marshal(PeerNode{IpAddress: ip, Port: port, IsActive: true})
	if err != nil {
		return fmt.Errorf("error marshaling add peer request body")
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "while creating request")
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error joining peers from %s", address))
	}
	response.Body.Close()
	return nil
}

func fetchBlocks(client *http.Client, peer PeerNode, hash database.Hash) ([]database.Block, error){
	var result SyncResult
	address := peer.SocketAddress()
	if !peer.IsActive {
		return result.Blocks, fmt.Errorf("%s not active", address)
	}
	url := fmt.Sprintf("%s://%s%s", "http", address, ApiRouteSync)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result.Blocks, errors.Wrap(err, "while creating request")
	}

	query := req.URL.Query()
	query.Set(ApiQueryParamAfter, hash.String())
	req.URL.RawQuery = query.Encode()

	response, err := client.Do(req)
	if err != nil {
		return result.Blocks, errors.Wrap(err, fmt.Sprintf("error fetching blocks from %s", address))
	}
	defer response.Body.Close()
	if err := readResponseJson(response, &result); err != nil {
		return result.Blocks, errors.Wrap(err, "error reading blocks in response")
	}
	return result.Blocks, nil
}
