package assets

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/akeil/scrapen/internal/pipeline"
)

var client = &http.Client{}

// DownloadImages finds img tags in the HTML and downloads the referenced images.
//
// Replaces the images src attribute with a local:// url.
func DownloadImages(ctx context.Context, i *pipeline.Item) (*pipeline.Item, error) {

	fetch := func(src string) (string, error) {
		src, err := resolveURL(src, i.URL)
		if err != nil {
			return "", err
		}

		req, err := http.NewRequestWithContext(ctx, "GET", src, nil)
		if err != nil {
			return "", err
		}

		fmt.Printf("Fetch image %q\n", src)
		res, err := client.Do(req)
		if err != nil {
			return "", err
		}

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
		err = i.PutAsset(id, contentType, data)
		if err != nil {
			return "", err
		}

		return id, nil
	}

	err := doImages(fetch, i)
	if err != nil {
		return nil, err
	}

	err = doMetadataImages(fetch, i)

	return i, err
}

type fetchFunc func(src string) (string, error)

func doImages(f fetchFunc, i *pipeline.Item) error {
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
	err := pipeline.WalkHTML(&b, i.HTML, handler)
	if err != nil {
		return err
	}

	i.HTML = b.String()
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

func doMetadataImages(f fetchFunc, i *pipeline.Item) error {
	if i.ImageURL == "" {
		return nil
	}

	href, err := resolveURL(i.ImageURL, i.URL)
	if err != nil {
		return err
	}

	id, err := f(href)
	if err != nil {
		return err
	}

	i.ImageURL = pipeline.StoreURL(id)

	return nil
}

func resolveURL(href, base string) (string, error) {
	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	ref, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	return b.ResolveReference(ref).String(), nil
}
