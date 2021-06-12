package content

import (
	"context"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

// ResolveURLs resolve any URLs in the content against the base URLs.
// Specifically, this will convert relative URLs into absolute ones so that
// they remain valid when the content is viewed offline or served from another
// host.
func ResolveURLs(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "content",
		"url":    t.ContentURL(),
	}).Info("Resolve URLs in content")

	if t.ImageURL != "" {
		u, err := t.ResolveURL(t.ImageURL)
		if err != nil {
			return err
		}
		t.ImageURL = u
	}

	doc := t.Document()
	base, err := url.Parse(t.ContentURL())
	if err != nil {
		return err
	}
	resolveContentURLs(doc, base)

	altDoc := t.AltDocument()
	altURL, err := url.Parse(t.AltURL)
	if altDoc != nil && err != nil {
		resolveContentURLs(altDoc, altURL)
	}

	return nil
}

var urlAttrs = []string{
	"src",
	"href",
	"srcset",
	"data-srcset",
	"data-src",
}

func resolveContentURLs(doc *goquery.Document, base *url.URL) {
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		for _, name := range urlAttrs {
			val, _ := s.Attr(name)
			if val != "" {
				newVal, err := resolveURL(base, val)
				if err == nil && val != newVal {
					s.RemoveAttr(name)
					s.SetAttr(name, newVal)
				}
			}
		}
	})
}

func resolveURL(base *url.URL, href string) (string, error) {
	h, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(h).String(), nil
}
