package rss

import (
	"context"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/akeil/scrapen/internal/pipeline"
)

func FindFeeds(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "rss",
	}).Info("Find feeds")

	t.Feeds = make([]pipeline.FeedInfo, 0)

	reader := func(tk html.Token) error {

		tt := tk.Type
		switch tt {
		case html.StartTagToken,
			html.SelfClosingTagToken:
			switch tk.DataAtom {
			case atom.Link:
				handleLink(tk, t)
			}
		case html.EndTagToken:
		}

		return nil
	}

	err := pipeline.ReadHTML(t.HTML, reader)
	if err != nil {
		log.WithFields(log.Fields{
			"task":   t.ID,
			"module": "rss",
			"error":  err,
		}).Warn("Failed to find feeds")

		return err
	}

	return nil
}

func handleLink(tk html.Token, t *pipeline.Task) {
	var (
		rel   string
		href  string
		title string
	)
	for _, attr := range tk.Attr {
		k := strings.ToLower(attr.Key)
		switch k {
		case "rel":
			rel = strings.ToLower(attr.Val)
		case "href":
			h, err := t.ResolveURL(attr.Val)
			if err != nil {
				log.WithFields(log.Fields{
					"task":   t.ID,
					"module": "rss",
					"error":  err,
				}).Warn("Could not resolve feed URL")
				return
			}
			href = h
		case "title":
			// TODO: unescape HTML
			title = attr.Val
		}
	}

	if rel != "alternate" || href == "" {
		return
	}

	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "rss",
		"rss":    href,
	}).Info("found link")

	t.Feeds = append(t.Feeds, pipeline.FeedInfo{
		URL:   href,
		Title: title,
	})
}
