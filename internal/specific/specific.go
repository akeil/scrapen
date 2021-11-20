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

	h, err := host(t.ContentURL())
	if err != nil {
		return err
	}

	switch h {
	case "bloomberg.com":
		bloomberg(t)
		break
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
	case "linkedin.com":
		return linkedin(ctx, t)
	}

	return nil
}

func host(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(u.Host, "www."), nil
}
