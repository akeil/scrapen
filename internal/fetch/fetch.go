package fetch

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

	setHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		return i, err
	}

	// TODO: use logger instead
	for k, v := range res.Request.Header {
		fmt.Printf("> %v: %v\n", k, v)
	}
	for k, v := range res.Header {
		fmt.Printf("< %v: %v\n", k, v)
	}

	i.StatusCode = res.StatusCode
	if res.Request.URL != nil {
		i.ActualURL = res.Request.URL.String()
	}

	err = errorFromStatus(res)
	if err != nil {
		return i, err
	}
	defer res.Body.Close()

	// decompress
	r, err := decompressed(res.Body, res.Header)
	if err != nil {
		return i, err
	}

	// decode charset
	s, err := readUTF8(r, res.Header)
	if err != nil {
		return i, err
	}
	i.HTML = s

	return i, nil
}

type browserProfile struct {
	UserAgent      string
	Accept         string
	AcceptLanguage string
}

var profiles = map[string]browserProfile{
	// Taken from Chromium on Linux
	"default": browserProfile{
		UserAgent:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36",
		Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		AcceptLanguage: "en-US,en;q=0.9,de;q=0.8",
	},
}

func setHeaders(req *http.Request) {
	profile := profiles["default"]
	req.Header.Add("User-Agent", profile.UserAgent)
	req.Header.Add("Accept", profile.Accept)
	req.Header.Add("Accept-Language", profile.AcceptLanguage)

	req.Header.Add("Accept-Encoding", supportedCompressions)

	//req.Header.Add("Connection", "keep-alive")

	// These seem to be relevant for bloomberg.com
	// (protected by Perimeterx Bot Defender)
	req.Header.Add("Path", req.URL.Path)
	req.Header.Add("Scheme", req.URL.Scheme)

}

func errorFromStatus(res *http.Response) error {
	// TODO: should we accept more status codes?
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("got HTTP status %v", res.StatusCode)
	}
	return nil
}
