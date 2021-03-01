package readable

import (
	"context"
	"strings"

	readability "github.com/go-shiori/go-readability"

	"github.com/akeil/scrapen/internal/pipeline"
)

func MakeReadable(ctx context.Context, i *pipeline.Item) (*pipeline.Item, error) {
	p := readability.NewParser()

	r := strings.NewReader(i.HTML)

	a, err := p.Parse(r, i.URL)
	if err != nil {
		return i, err
	}

	i.HTML = a.Content
	i.Title = a.Title

	return i, nil
}
