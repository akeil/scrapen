package fetch

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

var client = &http.Client{}

// Fetch fetches the HTML content for the given item.
func Fetch(ctx context.Context, t *pipeline.Task) error {

	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "fetch",
		"url":    t.ContentURL(),
	}).Info("Fetch content")

	html, err := fetchURL(ctx, t, t.URL)
	if err != nil {
		return err
	}

	// check for redirect from <meta http-equiv="refresh" ... />
	redirect, err := findRedirect(html)
	if err != nil {
		return err
	}
	if redirect != "" {
		log.WithFields(log.Fields{
			"task":   t.ID,
			"module": "fetch",
			"url":    redirect,
		}).Info("Redirect from <meta>")

		html, err = fetchURL(ctx, t, redirect)
		if err != nil {
			return err
		}
	}

	t.SetHTML(html)
	return nil
}

func fetchURL(ctx context.Context, t *pipeline.Task, url string) (string, error) {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "fetch",
		"url":    url,
	}).Info("Fetch URL")

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	setHeaders(req)

	for k, v := range req.Header {
		log.WithFields(log.Fields{
			"task":   t.ID,
			"module": "fetch",
			"url":    url,
			"header": k,
		}).Debug(fmt.Sprintf("Request Header: %v = %v", k, v))
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	for k, v := range res.Header {
		log.WithFields(log.Fields{
			"task":   t.ID,
			"module": "fetch",
			"url":    url,
			"header": k,
		}).Debug(fmt.Sprintf("Response Header: %v = %v", k, v))
	}

	t.StatusCode = res.StatusCode
	// TODO: does not seem to work in all cases...
	if res.Request.URL != nil {
		t.ActualURL = res.Request.URL.String()
	}

	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "fetch",
		"status": t.StatusCode,
		"url":    t.ActualURL,
	}).Info(fmt.Sprintf("Status %v", t.StatusCode))

	err = errorFromStatus(res)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	// decompress
	r, err := decompressed(t, res.Body, res.Header)
	if err != nil {
		return "", err
	}

	// decode charset
	s, err := readUTF8(t, r, res.Header)
	if err != nil {
		return "", err
	}

	return s, nil
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
	// Problem:
	// *some* URL shorteners will return a HTML site with a redirect
	// if they think the requests comes from a browser
	//
	// OTHERS will block requests if it does *not* look like a browser ...
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
