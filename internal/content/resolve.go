package content

import (
	"context"

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

	resolveContentURLs(t)

	return nil
}

var urlAttrs = []string{
	"src",
	"href",
}

func resolveContentURLs(t *pipeline.Task) {
	doc := t.Document()
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		for _, name := range urlAttrs {
			val, _ := s.Attr(name)
			if val != "" {
				newVal, err := t.ResolveURL(val)
				if err == nil {
					log.WithFields(log.Fields{
						"module": "content",
						"name":   name,
						"old":    val,
						"new":    newVal,
					}).Debug("Replacing URL")

					s.RemoveAttr(name)
					s.SetAttr(name, newVal)
				}
			}
		}
	})
}
