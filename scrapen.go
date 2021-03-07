package scrapen

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/assets"
	"github.com/akeil/scrapen/internal/content"
	"github.com/akeil/scrapen/internal/ebook"
	"github.com/akeil/scrapen/internal/fetch"
	"github.com/akeil/scrapen/internal/htm"
	"github.com/akeil/scrapen/internal/metadata"
	"github.com/akeil/scrapen/internal/pdf"
	"github.com/akeil/scrapen/internal/pipeline"
	"github.com/akeil/scrapen/internal/readable"
)

func Run(url string) error {
	//log.SetLevel(log.DebugLevel)
	log.SetLevel(log.InfoLevel)

	o := Options{
		Metadata:       true,
		Readability:    true,
		Clean:          true,
		DownloadImages: true,
	}
	s := pipeline.NewMemoryStore()
	id := uuid.New().String()
	result, err := doScrape(s, o, id, url)
	if err != nil {

		log.WithFields(log.Fields{
			"task":   id,
			"module": "main",
			"error":  err,
		}).Info("Scrape failed")

		return err
	}

	log.WithFields(log.Fields{
		"task":   result.ID,
		"module": "main",
		"url":    result.ContentURL(),
		"status": result.StatusCode,
	}).Info("Scrape complete")

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
	log.Info(fmt.Sprintf("Output to %q\n", outfile))

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

// Scrape creates and runs a scraping task for the given URL with given Options.
//
// The given Store will receive downloaded images and other assets.
func Scrape(s Store, o Options, id, url string) (Result, error) {
	t, err := doScrape(s, o, id, url)
	if err != nil {
		return Result{}, err
	}

	return resultFromTask(t), nil
}

func doScrape(s Store, o Options, id, url string) (*pipeline.Task, error) {
	ctx := context.Background()
	t := pipeline.NewTask(s, id, url)

	p := configurePipeline(o)
	err := p(ctx, t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// Options holds settings for a scraping task.
type Options struct {
	// Metadata controls whether metadata should be extracted.
	Metadata bool
	// Readability controls whether a readability script should be applied.
	Readability bool
	// Clean controls whether the resulting HTML should be stripped of unwanted tags.
	Clean bool
	// DownloadImages controls whether images from the content should be downloaded.
	DownloadImages bool
}

func configurePipeline(o Options) pipeline.Pipeline {
	p := []pipeline.Pipeline{
		fetch.Fetch,
	}

	if o.Metadata {
		p = append(p, metadata.ReadMetadata)
	}

	if o.Readability {
		p = append(p, readable.MakeReadable)
	}

	if o.Clean {
		p = append(p, content.Clean)
	}
	p = append(p, content.ResolveURLs)

	if o.DownloadImages {
		p = append(p, assets.DownloadImages)
	}
	return pipeline.BuildPipeline(p...)
}

// Store is the interface which receives downloaded image data.
//
// The Store is a simple key-value store.
// Each key stores the content type and the byte data.
type Store interface {
	// Put adds an entry to the store under the given key.
	Put(k, contentType string, data []byte) error
	// Get retrieves a store entry with the given key.
	Get(k string) (string, []byte, error)
}

// Result holds the result of a successful scraping task.
type Result struct {
	URL          string
	ActualURL    string
	CanonicalURL string
	StatusCode   int
	HTML         string
	Title        string
	Retrieved    time.Time
	Description  string
	PubDate      *time.Time
	Site         string
	SiteScheme   string
	Author       string
	ImageURL     string
}

func resultFromTask(t *pipeline.Task) Result {
	return Result{
		URL:          t.URL,
		ActualURL:    t.ActualURL,
		CanonicalURL: t.CanonicalURL,
		StatusCode:   t.StatusCode,
		HTML:         t.HTML,
		Title:        t.Title,
		Retrieved:    t.Retrieved,
		Description:  t.Description,
		PubDate:      t.PubDate,
		Site:         t.Site,
		SiteScheme:   t.SiteScheme,
		Author:       t.Author,
		ImageURL:     t.ImageURL,
	}
}

// ComposeFunc is used to compose an putput format for an item.
type ComposeFunc func(w io.Writer, i *pipeline.Task) error
