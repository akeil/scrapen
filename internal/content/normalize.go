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

	normalizeSpace(doc)

	return nil
}

// for each blocklevel element, eliminate any whitespace immediately before
// and after the tag.
//
// e.g.:    <h1> Abc </h1>  -->  <h1>Abc</h1>
func normalizeSpace(doc *goquery.Document) {
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {

	})
}

func isBlocklevel(tagName string) bool {
	return false
}
