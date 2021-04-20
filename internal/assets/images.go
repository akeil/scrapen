package assets

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/vincent-petithory/dataurl"

	"github.com/akeil/scrapen/internal/pipeline"
)

var client = &http.Client{}

// DownloadImages finds img tags in the HTML and downloads the referenced images.
//
// Replaces the images src attribute with a "store://xyz..." url.
func DownloadImages(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "assets",
	}).Info("Download images")

	f := func(src string) (string, error) {
		var i pipeline.ImageInfo
		var data []byte
		var err error
		if strings.HasPrefix(src, "data:") {
			i, data, err = fetchData(src)
		} else { // assume HTTP
			i, data, err = fetchHTTP(ctx, src)
		}

		if err != nil {
			return "", err
		}

		err = t.AddImage(i, data)
		if err != nil {

			log.WithFields(log.Fields{
				"task":   t.ID,
				"module": "assets",
				"error":  err,
			}).Warning("Failed to save image")

			return "", err
		}

		return i.ContentURL, nil
	}

	err := doImages(f, t)
	if err != nil {
		return err
	}

	return doMetadataImages(f, t)
}

type fetchFunc func(src string) (string, error)

func doImages(f fetchFunc, t *pipeline.Task) error {
	r := strings.NewReader(t.HTML)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var l sync.Lock

	doc.Selection.Find("img").Each(func(i int, s *goquery.Selection) {
		wg.Add(1)

		go func() {
			defer wg.Done()
			src, ok := s.Attr("src")
			if !ok || src == "" {
				// if do not understand how to download,
				// leave the image as is
				return
			}

			newSrc, err := f(src)
			if err != nil {
				// not much we can do about the error
				// we do not want to cancel the whole process
				// logging is sufficiently donw in fetch function
				return
			}
			l.Lock()
			s.SetAttr("src", newSrc)
			l.Unlock()
		}()
	})

	wg.Wait()

	html, err := doc.Selection.Find("body").First().Html()
	if err != nil {
		return err
	}
	t.HTML = html

	return nil
}

func fetchHTTP(ctx context.Context, src string) (pipeline.ImageInfo, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", src, nil)
	if err != nil {
		return pipeline.ImageInfo{}, nil, err
	}

	log.WithFields(log.Fields{
		"module": "assets",
		"url":    src,
	}).Info("Fetch image")

	res, err := client.Do(req)

	log.WithFields(log.Fields{
		"module": "assets",
		"url":    src,
		"status": res.StatusCode,
	}).Info("Got image response")

	if err != nil {
		return pipeline.ImageInfo{}, nil, err
	}

	if res.StatusCode != http.StatusOK {
		return pipeline.ImageInfo{}, nil, fmt.Errorf("got HTTP status %v", res.StatusCode)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return pipeline.ImageInfo{}, nil, err
	}

	// note: may be empty
	contentType := res.Header.Get("content-type")
	mime, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.WithFields(log.Fields{
			"module": "assets",
			"url":    src,
			"error":  err,
		}).Info(fmt.Sprintf("Failed to parse MIME type from %q", contentType))

		return pipeline.ImageInfo{}, nil, err
	}

	// note: may be empty for non-supported types.
	fileExt := fileExt(mime)

	id := uuid.New().String() + fileExt
	newSrc := pipeline.StoreURL(id)
	i := pipeline.ImageInfo{
		Key:         id,
		ContentURL:  newSrc,
		OriginalURL: src,
		ContentType: mime,
	}

	// in case there was a redirect on the image
	if res.Request.URL != nil {
		i.OriginalURL = res.Request.URL.String()
	}

	return i, data, nil
}

func fetchData(src string) (pipeline.ImageInfo, []byte, error) {
	d, err := dataurl.DecodeString(src)
	if err != nil {
		return pipeline.ImageInfo{}, nil, err
	}

	mime := d.MediaType.ContentType()

	// note: may be empty for non-supported types.
	fileExt := fileExt(mime)

	id := uuid.New().String() + fileExt
	newSrc := pipeline.StoreURL(id)

	i := pipeline.ImageInfo{
		Key:         id,
		ContentURL:  newSrc,
		OriginalURL: "",
		ContentType: mime,
	}

	return i, d.Data, nil
}

func doMetadataImages(f fetchFunc, t *pipeline.Task) error {
	if t.ImageURL == "" {
		return nil
	}

	// early exit
	// in case we already have the "main" image as part of the content
	existing := findExistingImage(t)
	if existing.ContentURL != "" {
		t.ImageURL = existing.ContentURL
		return nil
	}

	src, err := f(t.ImageURL)
	if err != nil {
		return err
	}

	t.ImageURL = src

	return nil
}

func findExistingImage(t *pipeline.Task) pipeline.ImageInfo {
	for _, img := range t.Images {
		if t.ImageURL == img.OriginalURL {
			return img
		}
	}
	return pipeline.ImageInfo{}
}
