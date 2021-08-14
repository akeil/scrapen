package main

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/htm"
	"github.com/akeil/scrapen/internal/pipeline"

	"github.com/akeil/scrapen"
)

func main() {

	url := os.Args[1]

	output := "./output"
	if len(os.Args) >= 3 {
		output = os.Args[2]
	}

	err := run(url, output)
	if err != nil {
		log.Fatal(err)
	}
}

type composeFunc func(w io.Writer, t *pipeline.Task) error

func run(url, output string) error {
	log.SetLevel(log.DebugLevel)
	//log.SetLevel(log.InfoLevel)
	s := pipeline.NewMemoryStore()
	o := &scrapen.Options{
		Metadata:       true,
		Readability:    true,
		Clean:          true,
		Normalize:      true,
		DownloadImages: true,
		FindFeeds:      true,
		SiteSpecific:   true,
		Store:          s,
	}
	a, err := scrapen.Scrape(url, o)
	if err != nil {
		return err
	}

	format := "html"
	outfile := fmt.Sprintf("%v.%v", output, format)
	log.Info(fmt.Sprintf("Output to %q\n", outfile))

	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()
	err = htm.Compose(f, taskFromArticle(a, s))
	if err != nil {
		return err
	}

	return nil
}

func taskFromArticle(a scrapen.Result, s scrapen.Store) *pipeline.Task {
	fs := make([]pipeline.FeedInfo, len(a.Feeds))
	for i, f := range a.Feeds {
		fs[i] = pipeline.FeedInfo{
			URL:   f.URL,
			Title: f.Title,
		}
	}

	imgs := make([]pipeline.ImageInfo, len(a.Images))
	for i, img := range a.Images {
		imgs[i] = pipeline.ImageInfo{
			Key:         img.Key,
			ContentURL:  img.ContentURL,
			ContentType: img.ContentType,
			OriginalURL: img.OriginalURL,
		}
	}

	encs := make([]pipeline.Enclosure, len(a.Enclosures))
	for i, e := range a.Enclosures {
		encs[i] = pipeline.Enclosure{
			Type:        e.Type,
			Title:       e.Title,
			URL:         e.URL,
			ContentType: e.ContentType,
			Description: e.Description,
		}
	}

	t := &pipeline.Task{
		URL:          a.URL,
		ActualURL:    a.ActualURL,
		CanonicalURL: a.CanonicalURL,
		StatusCode:   a.StatusCode,
		Title:        a.Title,
		Retrieved:    a.Retrieved,
		Description:  a.Description,
		PubDate:      a.PubDate,
		Site:         a.Site,
		SiteScheme:   a.SiteScheme,
		Author:       a.Author,
		ImageURL:     a.ImageURL,
		WordCount:    a.WordCount,
		Images:       imgs,
		Feeds:        fs,
		Enclosures:   encs,
		Store:        s,
	}
	t.SetHTML(a.HTML)
	return t
}
