package specific

import (
	"context"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

// SiteSpecific applies site-specific changes to the HTML content.
func SiteSpecific(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "specific",
	}).Info("Apply site-specific rules")

	u, err := url.Parse(t.ContentURL())
	if err != nil {
		return err
	}

	h := strings.TrimPrefix(u.Host, "www.")
	switch h {
	case "stackoverflow.com",
		"stackexchange.com",
		"superuser.com",
		"askubuntu.com",
		"ux.stackexchange.com",
		"unix.stackexchange.com",
		"datascience.stackexchange.com",
		"codereview.stackexchange.com":
		// TODO: too lazy to add them all:
		// https://stackexchange.com/sites
		stackoverflow(t)
		break
	}

	return nil
}
