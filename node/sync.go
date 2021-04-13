package node

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kparkins/yarbit/database"
	"github.com/pkg/errors"
)

func fetchBlocks(ctx context.Context, client *http.Client, address string, hash database.Hash) ([]database.Block, error) {
	var result SyncResult
	url := fmt.Sprintf("%s://%s%s", "http", address, ApiRouteSync)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
