package scrapen

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/akeil/scrapen/internal/assets"
	"github.com/akeil/scrapen/internal/clean"
	"github.com/akeil/scrapen/internal/ebook"
	"github.com/akeil/scrapen/internal/htm"
	"github.com/akeil/scrapen/internal/metadata"
	"github.com/akeil/scrapen/internal/pdf"
	"github.com/akeil/scrapen/internal/pipeline"
	"github.com/akeil/scrapen/internal/readable"
)

func Run(url string) error {
	result, err := Scrape(url)
	if err != nil {
		return err
	}

	format := "html"

	var compose ComposeFunc
	switch format {
	case "pdf":
		compose = pdf.Compose
	case "html":
		compose = htm.Compose
	case "epub":
		compose = ebook.Compose
	default:
		return fmt.Errorf("unsupported format: %q", format)
	}

	outfile := fmt.Sprintf("output.%v", format)
	fmt.Printf("Output to %q\n", outfile)

	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()
	err = compose(f, result)
	if err != nil {
		return err
	}

	return nil
}

func Scrape(url string) (*pipeline.Item, error) {
	ctx := context.Background()
	item := pipeline.NewItem(url)

	p := configurePipeline()
	result, err := p(ctx, item)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func configurePipeline() pipeline.Pipeline {
	// TODO: config and call options go here
	return pipeline.BuildPipeline(
		Fetch,
		metadata.ReadMetadata,
		readable.MakeReadable,
		clean.Clean,
		assets.DownloadImages)
}

// ComposeFunc is used to compose an putput format for an item.
type ComposeFunc func(w io.Writer, i *pipeline.Item) error
