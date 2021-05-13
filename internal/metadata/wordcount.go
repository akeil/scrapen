package metadata

import (
	"context"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

var pattern = regexp.MustCompile(`\w+`)

// CountWords adds the `WordCount` property to the the Task.
// It counts the number of all words in the content.
func CountWords(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "metadata",
		"url":    t.ContentURL(),
	}).Info("Count words")

	doc := t.Document()
	text := ""
	if doc != nil {
		text = doc.Selection.Find("body").First().Text()
	}

	words := pattern.FindAllString(text, -1)
	t.WordCount = len(words)

	return nil
}
