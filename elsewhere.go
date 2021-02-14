package elsewhere

import (
	"context"
	"fmt"
	"io"
	"os"

	"akeil.net/akeil/elsewhere/internal/assets"
	"akeil.net/akeil/elsewhere/internal/clean"
	"akeil.net/akeil/elsewhere/internal/ebook"
	"akeil.net/akeil/elsewhere/internal/htm"
	"akeil.net/akeil/elsewhere/internal/metadata"
	"akeil.net/akeil/elsewhere/internal/pdf"
	"akeil.net/akeil/elsewhere/internal/pipeline"
	"akeil.net/akeil/elsewhere/internal/readable"
)

func Run(url string) error {

	ctx := context.Background()
	item := pipeline.NewItem(url)

	p := configurePipeline()
	result, err := p(ctx, item)
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
