package scrapen

import (
	"context"
	"fmt"
	"net/http"

	"github.com/akeil/scrapen/internal/pipeline"
)

var client = &http.Client{}

// Fetch fetches the HTML content for the given item.
func Fetch(ctx context.Context, i *pipeline.Item) (*pipeline.Item, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", i.URL, nil)
	if err != nil {
		return i, err
	}

	res, err := client.Do(req)
	if err != nil {
		return i, err
	}

	err = errorFromStatus(res)
	if err != nil {
		return i, err
	}
	defer res.Body.Close()

	if res.Request.URL != nil {
		i.ActualURL = res.Request.URL.String()
	}

	s, err := readUTF8(res)
	if err != nil {
		return i, err
	}
	i.HTML = s

	return i, nil
}

func errorFromStatus(res *http.Response) error {
	// TODO: should we accept more status codes?
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("got HTTP status %v", res.StatusCode)
	}
	return nil
}
