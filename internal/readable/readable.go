package readable

import (
	"context"
	"strings"

	readability "github.com/go-shiori/go-readability"

	"github.com/akeil/scrapen/internal/pipeline"
)

func MakeReadable(ctx context.Context, t *pipeline.Task) error {
	p := readability.NewParser()

	r := strings.NewReader(t.HTML)

	a, err := p.Parse(r, t.URL)
	if err != nil {
		return err
	}

	t.HTML = a.Content
	t.Title = a.Title

	return nil
}
