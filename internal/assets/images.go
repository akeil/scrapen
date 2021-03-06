package assets

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/akeil/scrapen/internal/pipeline"
)

var client = &http.Client{}

// DownloadImages finds img tags in the HTML and downloads the referenced images.
//
// Replaces the images src attribute with a local:// url.
func DownloadImages(ctx context.Context, t *pipeline.Task) error {

	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "assets",
	}).Info("Download images")

	fetch := func(src string) (string, error) {
		req, err := http.NewRequestWithContext(ctx, "GET", src, nil)
		if err != nil {
			return "", err
		}

		log.WithFields(log.Fields{
			"task":   t.ID,
			"module": "assets",
			"url":    src,
		}).Info("Fetch image")

		res, err := client.Do(req)
		if err != nil {
			return "", err
		}

		log.WithFields(log.Fields{
			"task":   t.ID,
			"module": "assets",
			"url":    src,
			"status": res.StatusCode,
		}).Info("Got image response")

		if res.StatusCode != http.StatusOK {
			return "", fmt.Errorf("got HTTP status %v", res.StatusCode)
		}
		defer res.Body.Close()

		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}

		// note: may be empty
		contentType := res.Header.Get("content-type")

		id := uuid.New().String()
		err = t.PutAsset(id, contentType, data)
		if err != nil {

			log.WithFields(log.Fields{
				"task":   t.ID,
				"module": "assets",
				"error":  err,
			}).Warning("Failed to save image")

			return "", err
		}

		return id, nil
	}

	err := doImages(fetch, t)
	if err != nil {
		return err
	}

	return doMetadataImages(fetch, t)
}

type fetchFunc func(src string) (string, error)

func doImages(f fetchFunc, t *pipeline.Task) error {
	handler := func(t html.Token, w io.StringWriter) (bool, error) {
		if t.DataAtom != atom.Img {
			return false, nil
		}

		var err error
		tt := t.Type
		switch tt {
		case html.StartTagToken:
			w.WriteString("<")
			w.WriteString(t.Data)
			err = localImage(t.Attr, f, w)
			if err != nil {
				return false, err
			}
			w.WriteString(">")
			return true, nil
		case html.SelfClosingTagToken:
			w.WriteString("<")
			w.WriteString(t.Data)
			err = localImage(t.Attr, f, w)
			if err != nil {
				return false, err
			}
			w.WriteString("/>")
			return true, nil
		}

		return false, nil
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
				Val:       pipeline.StoreURL(newSrc),
			}, w)
		} else {
			pipeline.WriteAttr(attr, w)
		}
	}
	return nil
}

func doMetadataImages(f fetchFunc, t *pipeline.Task) error {
	if t.ImageURL == "" {
		return nil
	}

	id, err := f(t.ImageURL)
	if err != nil {
		return err
	}

	t.ImageURL = pipeline.StoreURL(id)

	return nil
}
