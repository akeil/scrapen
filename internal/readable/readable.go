package readable

import (
	"context"
	"strings"

	readability "github.com/go-shiori/go-readability"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

func MakeReadable(ctx context.Context, t *pipeline.Task) error {

	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "readable",
		"url":    t.ContentURL(),
	}).Info("Apply readability")

	p := readability.NewParser()
	r := strings.NewReader(t.HTML())

	a, err := p.Parse(r, t.URL)
	if err != nil {
		return err
	}

	t.SetHTML(a.Content)
	t.Title = a.Title

	return nil
}
