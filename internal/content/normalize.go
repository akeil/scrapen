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

	r := strings.NewReader(t.HTML)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}

	err = normalizeSpace(doc)
	if err != nil {
		return err
	}

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
