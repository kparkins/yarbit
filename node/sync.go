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
		peerAddress := peer.SocketAddress()
		status, err := fetchPeerStatus(client, peerAddress)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v while checking status of %s\n", err, peer.SocketAddress())
			n.RemovePeer(peer)
			continue
		}
		n.AddPeers(status.KnownPeers)
		if err := joinPeers(client, peerAddress, n.config.IpAddress, n.config.Port); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		if status.Number <= n.LatestBlockNumber() {
			continue
		}
		blocks, err := fetchBlocks(client, peerAddress, n.LatestBlockHash())
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

func fetchPeerStatus(client *http.Client, address string) (StatusResponse, error) {
	var status StatusResponse
	url := fmt.Sprintf("%s://%s%s", "http", address, ApiRouteStatus)
	response, err := client.Get(url)
	if err != nil {
		return status, errors.Wrap(err, fmt.Sprintf("error fetching peers from %s", address))
	}
	statusResponse := StatusResponse{}
	if err := readJsonResponse(response, &statusResponse); err != nil {
		return status, err
	}
	return statusResponse, nil
}

func joinPeers(client *http.Client, address, ip string, port uint64) error {
	url := fmt.Sprintf("%s://%s%s", "http", address, ApiRouteAddPeer)
	body, err := json.Marshal(PeerNode{IpAddress: ip, Port: port, IsActive: true})
	if err != nil {
		return fmt.Errorf("error marshaling add peer request body")
	}
	response, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error joining peers from %s", address))
	}
	_ = response.Body.Close()
	return nil
}

func fetchBlocks(client *http.Client, address string, hash database.Hash) ([]database.Block, error){
	var result SyncResult
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

	if err := readJsonResponse(response, &result); err != nil {
		return result.Blocks, errors.Wrap(err, "error reading blocks in response")
	}
	return result.Blocks, nil
}
