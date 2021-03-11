package main

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/ebook"
	"github.com/akeil/scrapen/internal/htm"
	"github.com/akeil/scrapen/internal/pdf"
	"github.com/akeil/scrapen/internal/pipeline"

	"github.com/akeil/scrapen"
)

func main() {

	url := os.Args[1]

	err := run(url)
	if err != nil {
		log.Fatal(err)
	}
}

type composeFunc func(w io.Writer, t *pipeline.Task) error

func run(url string) error {
	//log.SetLevel(log.DebugLevel)
	log.SetLevel(log.InfoLevel)
	s := pipeline.NewMemoryStore()
	o := &scrapen.Options{
		Metadata:       true,
		Readability:    true,
		Clean:          true,
		DownloadImages: true,
		FindFeeds:      true,
		Store:          s,
	}
	a, err := scrapen.Scrape(url, o)
	if err != nil {
		return err
	}

	format := "html"

	var compose composeFunc
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
	log.Info(fmt.Sprintf("Output to %q\n", outfile))

	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()
	err = compose(f, taskFromArticle(a, s))
	if err != nil {
		return err
	}

	return nil
}

func taskFromArticle(a scrapen.Result, s scrapen.Store) *pipeline.Task {
	return &pipeline.Task{
		URL:          a.URL,
		ActualURL:    a.ActualURL,
		CanonicalURL: a.CanonicalURL,
		StatusCode:   a.StatusCode,
		HTML:         a.HTML,
		Title:        a.Title,
		Retrieved:    a.Retrieved,
		Description:  a.Description,
		PubDate:      a.PubDate,
		Site:         a.Site,
		SiteScheme:   a.SiteScheme,
		Author:       a.Author,
		ImageURL:     a.ImageURL,
		Store:        s,
	}
}
