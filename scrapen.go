package scrapen

import (
	"context"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/assets"
	"github.com/akeil/scrapen/internal/content"
	//	"github.com/akeil/scrapen/internal/ebook"
	"github.com/akeil/scrapen/internal/fetch"
	//	"github.com/akeil/scrapen/internal/htm"
	"github.com/akeil/scrapen/internal/metadata"
	//	"github.com/akeil/scrapen/internal/pdf"
	"github.com/akeil/scrapen/internal/pipeline"
	"github.com/akeil/scrapen/internal/readable"
	"github.com/akeil/scrapen/internal/rss"
)

// Scrape creates and runs a scraping task for the given URL with given Options.
//
// The given Store will receive downloaded images and other assets.
func Scrape(url string, o *Options) (Result, error) {
	t, err := doScrape(url, o)
	if err != nil {
		return Result{}, err
	}

	return resultFromTask(t), nil
}

func doScrape(url string, o *Options) (*pipeline.Task, error) {
	if o == nil {
		o = DefaultOptions()
	}
	ctx := context.Background()
	id := uuid.New().String()
	t := pipeline.NewTask(o.Store, id, url)

	p := configurePipeline(o)
	err := p(ctx, t)
	if err != nil {

		log.WithFields(log.Fields{
			"task":   t.ID,
			"module": "main",
			"error":  err,
		}).Warn("Scrape failed")

		return nil, err
	}

	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "main",
		"url":    t.ContentURL(),
		"status": t.StatusCode,
	}).Info("Scrape complete")

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
	// Detect RSS feeds
	FindFeeds bool
	// A Store is required if DownloadImages is true.
	Store Store
}

// DefaultOptions creates default scrape settings.
func DefaultOptions() *Options {
	return &Options{
		Metadata:       true,
		Readability:    true,
		Clean:          true,
		DownloadImages: false,
		FindFeeds:      false,
		Store:          nil,
	}
}

func configurePipeline(o *Options) pipeline.Pipeline {
	p := []pipeline.Pipeline{
		fetch.Fetch,
	}

	if o.Metadata {
		p = append(p, metadata.ReadMetadata)
	}

	if o.FindFeeds {
		p = append(p, rss.FindFeeds)
	}

	if o.Readability {
		p = append(p, readable.MakeReadable)
	}

	if o.Clean {
		p = append(p, content.Clean)
	}
	p = append(p, content.ResolveURLs)

	// we should call this AFTER modifiying the HTML
	p = append(p, content.Sanitize)

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
	Feeds        []pipeline.FeedInfo
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
		Feeds:        t.Feeds,
		ImageURL:     t.ImageURL,
	}
}
