package content

import (
	"context"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

func Normalize(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "content",
	}).Info("Normalize HTML")

	doc, err := t.Document()
	if err != nil {
		return err
	}

	err = normalizeSpace(doc)
	if err != nil {
		return err
	}

	deduplicateImage(t, doc)
	deduplicateTitle(doc, t.Title)

	return nil
}

// for each blocklevel element, eliminate any whitespace immediately before
// and after the tag.
//
// e.g.:    <h1> Abc </h1>  -->  <h1>Abc</h1>
func normalizeSpace(doc *goquery.Document) error {
	var err error
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		if isBlocklevel(tag) {
			h, e := s.Html()
			if e != nil {
				err = e
				return
			}
			s.SetHtml(strings.TrimSpace(h))
		}
	})

	return err
}

func deduplicateTitle(doc *goquery.Document, title string) {
	doc.Selection.Find("h1, h2, h3, h4, h5, h6").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		log.Info(tag)
		t := strings.TrimSpace(s.Text())
		if strings.EqualFold(t, title) {
			s.Remove()
		}
	})
}

// If the same image URL that is the "main" image for the article
// also appears in the content, remove it from content.
func deduplicateImage(t *pipeline.Task, doc *goquery.Document) {
	if t.ImageURL == "" {
		return
	}

	doc.Selection.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if !exists {
			return
		}

		if src == t.ImageURL {
			s.Remove()
		}
	})
}
