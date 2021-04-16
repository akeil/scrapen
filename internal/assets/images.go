package assets

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/vincent-petithory/dataurl"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

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

	fetch := func(src string) (string, error) {
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

	err := doImages(fetch, t)
	if err != nil {
		return err
	}

	return doMetadataImages(fetch, t)
}

type fetchFunc func(src string) (string, error)

func doImages(f fetchFunc, t *pipeline.Task) error {
	handler := func(tk html.Token, w io.StringWriter) (bool, error) {
		if tk.DataAtom != atom.Img {
			return false, nil
		}

		// TODO: account for duplicates
		// i.e. if we already have the image, re-use it

		var err error
		var tmp strings.Builder
		tt := tk.Type
		switch tt {
		case html.StartTagToken:
			tmp.WriteString("<")
			tmp.WriteString(tk.Data)
			err = localImage(tk.Attr, f, &tmp)
			tmp.WriteString(">")
		case html.SelfClosingTagToken:
			tmp.WriteString("<")
			tmp.WriteString(tk.Data)
			err = localImage(tk.Attr, f, &tmp)
			tmp.WriteString("/>")
		default:
			// should not be possible
			return false, nil
		}

		// if we encounter a download error,
		// leave the image as is.
		if err != nil {
			log.WithFields(log.Fields{
				"task":   t.ID,
				"module": "assets",
				"error":  err,
			}).Warning("Failed to download image")

			return false, nil
		}
		w.WriteString(tmp.String())
		return true, nil
	}

	var b strings.Builder
	err := pipeline.WalkHTML(&b, t.HTML, handler)
	if err != nil {
		return err
	}

	t.HTML = b.String()
	return nil
}

func localImage(a []html.Attribute, f fetchFunc, w io.StringWriter) error {
	for _, attr := range a {
		if attr.Key == "src" {
			newSrc, err := f(attr.Val)
			if err != nil {
				return err
			}
			pipeline.WriteAttr(html.Attribute{
				Namespace: "",
				Key:       "src",
				Val:       newSrc,
			}, w)
		} else {
			pipeline.WriteAttr(attr, w)
		}
	}
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
