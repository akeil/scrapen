package elsewhere

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"akeil.net/akeil/elsewhere/internal/pipeline"
)

var client = &http.Client{}

func Fetch(ctx context.Context, i *pipeline.Item) (*pipeline.Item, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", i.Url, nil)
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

	// TODO: would be nice if we could set the resolved URL

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return i, err
	}

	i.Html = string(data)

	return i, nil
}

func errorFromStatus(res *http.Response) error {
	// TODO: should we accept more status codes?
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("got HTTP status %v", res.StatusCode)
	}
	return nil
}
