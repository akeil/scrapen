package content

import (
	"context"

	"github.com/microcosm-cc/bluemonday"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

// Sanitize removes malicious content from the HTML document.
// This should be called after all other modifications have been performed.
func Sanitize(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "content",
	}).Info("Sanitize HTML")

	//p := bluemonday.UGCPolicy()
	p := createPolicy()

	t.SetHTML(p.Sanitize(t.HTML()))

	return nil
}

func createPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements(whitelist...)

	m := make(map[string][]string)
	for elem, a := range attrWhitelist {
		for _, attr := range a {
			_, ok := m[attr]
			if !ok {
				m[attr] = make([]string, 0)
			}
			m[attr] = append(m[attr], elem)
		}
	}

	for attr, elems := range m {
		p.AllowAttrs(attr).OnElements(elems...)
	}

	return p
}
