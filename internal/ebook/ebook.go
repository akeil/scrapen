package ebook

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/url"
	"os"
	"strings"

	"github.com/bmaupin/go-epub"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/akeil/scrapen/internal/pipeline"
)

// Compose creates an EPUB file from the given item.
func Compose(w io.Writer, i *pipeline.Item) error {
	tempdir, err := ioutil.TempDir("", "ebook-*")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(tempdir)
	}()

	e := epub.NewEpub(i.Title)
	err = composeEPUB(e, i, tempdir)
	if err != nil {
		return err
	}

	// TODO
	// epub library only allows writing to dst path
	// so we write to a temp file
	// and then copy the temp file to the actual dst writer
	tmp, err := ioutil.TempFile("", "*.epub")
	if err != nil {
		return err
	}
	defer tmp.Close()
	defer func() {
		os.Remove(tmp.Name())
	}()

	err = e.Write(tmp.Name())
	if err != nil {
		return err
	}

	_, err = io.Copy(w, tmp)
	return err
}

func composeEPUB(e *epub.Epub, i *pipeline.Item, tempdir string) error {
	e.SetTitle(i.Title)

	html, err := prepareContent(e, i, tempdir)
	if err != nil {
		return err
	}
	e.AddSection(html, i.Title, "", "")

	return nil
}

// prepareContent builds the HTML that is to be included in the EPUB.
// It replaces references to images to internal references within the epub file.
func prepareContent(e *epub.Epub, i *pipeline.Item, tempdir string) (string, error) {
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
			err = prepareImage(t.Attr, i, w, e, tempdir)
			if err != nil {
				return false, err
			}
			w.WriteString(">")
			return true, nil
		case html.SelfClosingTagToken:
			w.WriteString("<")
			w.WriteString(t.Data)
			err = prepareImage(t.Attr, i, w, e, tempdir)
			if err != nil {
				return false, err
			}
			w.WriteString("/>")
			return true, nil
		}

		return false, nil
	}

	var b strings.Builder
	err := pipeline.WalkHTML(&b, i.Html, handler)
	return b.String(), err
}

func prepareImage(a []html.Attribute, i *pipeline.Item, w io.StringWriter, e *epub.Epub, tempdir string) error {

	for _, attr := range a {
		if attr.Key == "src" {
			u, err := url.Parse(attr.Val)
			if err != nil {
				return err
			}
			if u.Scheme == "local" {
				id := u.Host
				err = addImage(tempdir, id, i, w, e)
				if err != nil {
					return nil
				}
			} else {
				pipeline.WriteAttr(attr, w)
			}
		} else {
			pipeline.WriteAttr(attr, w)
		}
	}

	return nil
}

func addImage(tempdir, id string, i *pipeline.Item, w io.StringWriter, e *epub.Epub) error {
	contentType, data, err := i.GetAsset(id)
	if err != nil {
		return err
	}

	ext, err := fileExtByMime(contentType)
	if err != nil {
		return err
	}

	// We can only add *paths* to images, not image data.
	// Therfore, we need to write to a temp directory and add the path.
	// Images Will be read when the epub is written.
	f, err := ioutil.TempFile(tempdir, id+"*"+ext)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	filename := id + ext

	src, err := e.AddImage(f.Name(), filename)
	if err != nil {
		return err
	}

	pipeline.WriteAttr(html.Attribute{
		Key: "src",
		Val: src,
	}, w)

	return nil
}

func fileExtByMime(contentType string) (string, error) {
	exts, err := mime.ExtensionsByType(contentType)
	if err != nil {
		return "", err
	}
	if len(exts) == 0 {
		return "", fmt.Errorf("found no suitable file extension for content-type %q", contentType)
	}

	// exts may contain multiple alternatives.
	// Some of these (e.g. 'jfif' for image/jpeg) are not understood by then
	// epub package. So we need some control over which extension is used.
	ext := selectExt(exts)
	return ext, nil
}

var preferredExts = []string{
	".jpg",
	".png",
	".gif",
}

func selectExt(exts []string) string {
	for _, ext := range exts {
		for _, pref := range preferredExts {
			if ext == pref {
				return pref
			}
		}
	}
	// default, first entry
	return exts[0]
}
