package fetch

import (
	"context"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/http/cookiejar"

	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

// Fetch fetches the HTML content for the given item.
func Fetch(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "fetch",
		"url":    t.ContentURL(),
	}).Info("Fetch content")

	opts := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
	jar, err := cookiejar.New(opts)
	if err != nil {
		return err
	}
	client := &http.Client{
		Jar: jar,
	}

	actURL, html, err := fetchURL(ctx, client, t, t.URL)
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

		actURL, html, err = fetchURL(ctx, client, t, redirect)
		if err != nil {
			return err
		}
	}

	// If the AMP URL was supplied, we want to fetch the canonical document
	// for additional metadata.
	// If the "normal" URL was supplied, we want to fetch the AMP document
	// for better content extraction.
	isAMP, canonicalURL := checkAMP(html)
	if isAMP {
		log.WithFields(log.Fields{
			"task":      t.ID,
			"module":    "fetch",
			"url":       actURL,
			"canonical": canonicalURL,
		}).Debug("Initial URL is AMP, fetching canonical")

		t.SetAltHTML(html)
		t.AltURL = actURL

		actURL, html, err = fetchURL(ctx, client, t, canonicalURL)
		if err != nil {
			return err
		}
		t.SetHTML(html)
		t.ActualURL = actURL

	} else {
		t.SetHTML(html)
		t.ActualURL = actURL

		// If an AMP (https://amp.dev/) version is available, fetch that.
		// Often easier to make readable.
		amp := findAmpUrl(html)
		if amp != "" {
			err = fetchAMP(ctx, client, t, amp)
			if err != nil {
				log.WithFields(log.Fields{
					"task":   t.ID,
					"module": "fetch",
					"url":    amp,
				}).Info("Failed to fetch AMP version")
			}
		}
	}

	return nil
}

func fetchAMP(ctx context.Context, client *http.Client, t *pipeline.Task, url string) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "fetch",
		"url":    url,
	}).Info("Fetch AMP version")
	url, err := t.ResolveURL(url)
	if err != nil {
		return err
	}

	actURL, html, err := fetchURL(ctx, client, t, url)
	if err != nil {
		return err
	}

	t.SetAltHTML(html)
	t.AltURL = actURL
	return nil
}

func fetchURL(ctx context.Context, client *http.Client, t *pipeline.Task, url string) (string, string, error) {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "fetch",
		"url":    url,
	}).Info("Fetch URL")

	var actURL string

	res, err := doRequest(ctx, client, url)
	if err != nil {
		return "", "", err
	}

	// if failed, repeat the request with cookies
	if res.StatusCode != http.StatusOK {
		if didReceiveCookie(res) {
			log.WithFields(log.Fields{
				"task":   t.ID,
				"module": "fetch",
				"url":    url,
				"status": res.StatusCode,
			}).Info("Repeat request with cookies")
			res, err = doRequest(ctx, client, url)
			if err != nil {
				return "", "", err
			}
		}
	}

	t.StatusCode = res.StatusCode
	if res.Request.URL != nil {
		actURL = res.Request.URL.String()
	}

	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "fetch",
		"status": t.StatusCode,
		"url":    t.ActualURL,
	}).Info(fmt.Sprintf("Status %v", t.StatusCode))

	err = errorFromStatus(res)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	// decompress
	r, err := decompressed(t, res.Body, res.Header)
	if err != nil {
		return "", "", err
	}

	// decode charset
	s, err := readUTF8(t, r, res.Header)
	if err != nil {
		return "", "", err
	}

	return actURL, s, nil
}

func doRequest(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add cookies collected in previous requests
	log.Info(fmt.Sprintf("Set cookies for %v", req.URL))
	for _, c := range client.Jar.Cookies(req.URL) {
		log.Info(fmt.Sprintf("Cookie: %v", c.String()))
		req.AddCookie(c)
	}

	setHeaders(req)

	for k, v := range req.Header {
		log.WithFields(log.Fields{
			"module": "fetch",
			"url":    url,
			"header": k,
		}).Debug(fmt.Sprintf("Header value: %v", v))
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	for k, v := range res.Header {
		log.WithFields(log.Fields{
			"module": "fetch",
			"url":    url,
			"header": k,
		}).Debug(fmt.Sprintf("Response Header: %v = %v", k, v))
	}

	// Set cookies for all subsequent requests
	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return res, nil
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

	// These seem to be relevant for bloomberg.com
	// (protected by Perimeterx Bot Defender)
	req.Header.Add("Path", req.URL.Path)
	req.Header.Add("Scheme", req.URL.Scheme)

	// might be relevant for other "bot detection" mechanisms
	req.Header.Add("Authority", req.URL.Host)
	req.Header.Add("Method", req.Method)
	req.Header.Add("Connection", "keep-alive")
}

func didReceiveCookie(res *http.Response) bool {
	c := res.Header.Get("Set-Cookie")
	return c != ""
}

func errorFromStatus(res *http.Response) error {
	// TODO: should we accept more status codes?
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("got HTTP status %v", res.StatusCode)
	}
	return nil
}
