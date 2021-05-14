package scrapen

import (
	"context"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/assets"
	"github.com/akeil/scrapen/internal/content"
	"github.com/akeil/scrapen/internal/fetch"
	//	"github.com/akeil/scrapen/internal/htm"
	"github.com/akeil/scrapen/internal/metadata"
	"github.com/akeil/scrapen/internal/pipeline"
	"github.com/akeil/scrapen/internal/readable"
	"github.com/akeil/scrapen/internal/rss"
	"github.com/akeil/scrapen/internal/specific"
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
	// Normalize controls whether we should attempt to normalize the HTML content,
	// e.g. remove unnecessary whitespace.
	Normalize bool
	// DownloadImages controls whether images from the content should be downloaded.
	DownloadImages bool
	// SiteSpecific controls whether to apply site-specific content-selectors.
	SiteSpecific bool
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
		Normalize:      true,
		DownloadImages: false,
		SiteSpecific:   false,
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
		p = append(p, metadata.FallbackImage)
	}

	if o.FindFeeds {
		p = append(p, rss.FindFeeds)
	}

	if o.SiteSpecific {
		p = append(p, specific.SiteSpecific)
	}

	p = append(p, content.Prepare)

	// Do this *before* Readability
	p = append(p, content.ResolveURLs)
	if o.Readability {
		p = append(p, readable.MakeReadable)
	}

	if o.Clean {
		p = append(p, content.Clean)
	}
	if o.Normalize {
		p = append(p, content.Normalize)
	}

	// we should call this AFTER modifiying the HTML
	p = append(p, content.Sanitize)

	// working on the final content HTML
	p = append(p, metadata.CountWords)
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
	WordCount    int
	Feeds        []Feed
	Images       []Image
	Enclosures   []Enclosure
	ImageURL     string
}

type Feed struct {
	URL   string
	Title string
}

type Image struct {
	Key         string
	ContentURL  string
	ContentType string
	OriginalURL string
}

type Enclosure struct {
	Type        string
	Title       string
	URL         string
	ContentType string
	Description string
}

func resultFromTask(t *pipeline.Task) Result {
	fs := make([]Feed, len(t.Feeds))
	for i, fi := range t.Feeds {
		fs[i] = Feed{
			URL:   fi.URL,
			Title: fi.Title,
		}
	}

	imgs := make([]Image, len(t.Images))
	for i, img := range t.Images {
		imgs[i] = Image{
			Key:         img.Key,
			ContentURL:  img.ContentURL,
			ContentType: img.ContentType,
			OriginalURL: img.OriginalURL,
		}
	}

	encs := make([]Enclosure, len(t.Enclosures))
	for i, e := range t.Enclosures {
		encs[i] = Enclosure{
			Type:        e.Type,
			Title:       e.Title,
			URL:         e.URL,
			ContentType: e.ContentType,
			Description: e.Description,
		}
	}

	doc := t.Document()
	html := ""
	if doc != nil {
		html, _ = doc.Selection.Find("body").First().Html()
	}

	return Result{
		URL:          t.URL,
		ActualURL:    t.ActualURL,
		CanonicalURL: t.CanonicalURL,
		StatusCode:   t.StatusCode,
		HTML:         html,
		Title:        t.Title,
		Retrieved:    t.Retrieved,
		Description:  t.Description,
		PubDate:      t.PubDate,
		Site:         t.Site,
		SiteScheme:   t.SiteScheme,
		Author:       t.Author,
		WordCount:    t.WordCount,
		Feeds:        fs,
		Images:       imgs,
		Enclosures:   encs,
		ImageURL:     t.ImageURL,
	}
}
