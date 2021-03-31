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

func syncWithPeers(ctx context.Context, n *Node) {
	knownPeers := n.Peers()
	client := &http.Client{
		Timeout: 4 * time.Second,
	}
	fmt.Println("Sync with peers")
	nodeAddress := fmt.Sprintf("%s:%d", n.config.IpAddress, n.config.Port)
	for _, peer := range knownPeers {
		fmt.Printf("peer %s\n", peer.SocketAddress())
		peerAddress := peer.SocketAddress()
		status, err := fetchPeerStatus(ctx, client, peerAddress)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v while checking status of %s\n", err, peer.SocketAddress())
			n.RemovePeer(peer)
			continue
		}
		fmt.Printf("got status %v\n", status)
		status.KnownPeers = FilterPeers(status.KnownPeers, func(s string) bool {
			return s != nodeAddress
		})
		n.AddPeers(status.KnownPeers)
		n.AddPendingTxs(status.PendingTxs)
		if err := joinPeers(ctx, client, peerAddress, n.config.IpAddress, n.config.Port); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		fmt.Printf("latest block num %d\n", n.LatestBlockNumber())
		if status.Number < n.LatestBlockNumber() {
			continue
		}
		if status.Number == 0 && !n.LatestBlockHash().IsEmpty() {
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

func joinPeers(ctx context.Context, client *http.Client, address, ip string, port uint64) error {
	url := fmt.Sprintf("%s://%s%s", "http", address, ApiRouteAddPeer)
	body, err := json.Marshal(PeerNode{IpAddress: ip, Port: port, IsActive: true})
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
		return errors.Wrap(err, fmt.Sprintf("error joining peers from %s", address))
	}
	_ = response.Body.Close()
	return nil
}

func fetchBlocks(ctx context.Context, client *http.Client, address string, hash database.Hash) ([]database.Block, error) {
	var result SyncResult
	url := fmt.Sprintf("%s://%s%s", "http", address, ApiRouteSync)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return result.Blocks, errors.Wrap(err, "while creating request")
	}
	fmt.Println("Fetching blocks")
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
	fmt.Printf("got blocks %v\n", result.Blocks)
	return result.Blocks, nil
}