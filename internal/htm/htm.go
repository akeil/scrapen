package htm

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	_ "embed"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/akeil/scrapen/internal/pipeline"
)

//go:embed style.css
var style string

func Compose(w io.Writer, t *pipeline.Task) error {
	var b strings.Builder

	b.WriteString("<html>")
	writeHead(&b, t)
	err := writeBody(&b, t)
	if err != nil {
		return err
	}
	b.WriteString("</html>")

	_, err = w.Write([]byte(b.String()))
	if err != nil {
		return err
	}
	return nil
}

func writeHead(b *strings.Builder, t *pipeline.Task) {
	b.WriteString("<head>")
	b.WriteString("<meta charset=\"utf-8\"/>")
	b.WriteString(fmt.Sprintf("<title>%v</title>", html.EscapeString(t.Title)))
	b.WriteString("<style>")
	b.WriteString(style)
	b.WriteString("</style>")
	b.WriteString("</head>")
}

func writeBody(b *strings.Builder, t *pipeline.Task) error {
	b.WriteString("<body>")
	err := writeMetadata(b, t)
	if err != nil {
		return err
	}
	err = writeContent(b, t)
	if err != nil {
		return err
	}
	writeFooter(b, t)
	b.WriteString("</body>")
	return nil
}

func writeMetadata(b *strings.Builder, t *pipeline.Task) error {
	// TODO: include <h1> with title?
	b.WriteString("<header>")
	if t.ImageURL != "" {
		attr := []html.Attribute{html.Attribute{Key: "src", Val: t.ImageURL}}
		b.WriteString("<img ")
		dataImage(attr, t, b)
		b.WriteString("/>")
	}

	if t.Title != "" {
		b.WriteString("<h1>")
		b.WriteString(t.Title)
		b.WriteString("</h1>")
	}

	if t.Author != "" {
		b.WriteString("<p>")
		b.WriteString("By ")
		b.WriteString("<strong>")
		b.WriteString(t.Author)
		b.WriteString("</strong>")
		b.WriteString("</p>")
	}

	if t.PubDate != nil {
		b.WriteString("<p>")
		b.WriteString("Published ")
		b.WriteString("<time datetime=\"")
		b.WriteString(t.PubDate.Format(time.RFC3339))
		b.WriteString("\">")
		b.WriteString(t.PubDate.Local().Format(time.ANSIC))
		b.WriteString("</time>")
		b.WriteString("</p>")
	}

	if t.Site != "" {
		b.WriteString("<p><a href=\"")
		b.WriteString(t.SiteScheme)
		b.WriteString("://")
		b.WriteString(t.Site)
		b.WriteString("\">")
		b.WriteString(t.Site)
		b.WriteString("</a></p>")
	}

	if t.Description != "" {
		b.WriteString("<p>")
		b.WriteString(t.Description)
		b.WriteString("</p>")
	}

	b.WriteString("</header>")
	return nil
}

func writeContent(b *strings.Builder, t *pipeline.Task) error {
	handler := func(tk html.Token, w io.StringWriter) (bool, error) {
		if tk.DataAtom != atom.Img {
			return false, nil
		}

		var err error
		tt := tk.Type
		switch tt {
		case html.StartTagToken:
			w.WriteString("<")
			w.WriteString(tk.Data)
			err = dataImage(tk.Attr, t, w)
			if err != nil {
				return false, err
			}
			w.WriteString(">")
			return true, nil
		case html.SelfClosingTagToken:
			w.WriteString("<")
			w.WriteString(tk.Data)
			err = dataImage(tk.Attr, t, w)
			if err != nil {
				return false, err
			}
			w.WriteString("/>")
			return true, nil
		}

		return false, nil
	}

	return pipeline.WalkHTML(b, t.HTML, handler)
}

func dataImage(a []html.Attribute, t *pipeline.Task, w io.StringWriter) error {
	for _, attr := range a {
		if attr.Key == "src" {
			storeID := pipeline.ParseStoreID(attr.Val)
			if storeID != "" {
				contentType, data, err := t.GetAsset(storeID)
				if err != nil {
					return err
				}
				enc := base64.StdEncoding.EncodeToString(data)
				v := "data:"
				// TODO: contentType
				v += contentType + ";"
				v += "base64,"
				v += enc
				pipeline.WriteAttr(html.Attribute{
					Key: "src",
					Val: v,
				}, w)

			} else {
				pipeline.WriteAttr(attr, w)
			}
		} else {
			pipeline.WriteAttr(attr, w)
		}
	}
	return nil
}

func writeFooter(b *strings.Builder, t *pipeline.Task) {
	b.WriteString("<footer>")
	b.WriteString("<p>")

	b.WriteString("Retrieved on ")
	// see:
	// http://microformats.org/wiki/datetime-design-pattern
	b.WriteString("<time datetime=\"")
	b.WriteString(t.Retrieved.Format(time.RFC3339))
	b.WriteString("\">")
	b.WriteString(t.Retrieved.Local().Format(time.ANSIC))
	b.WriteString("</time>")
	b.WriteString(" | ")

	b.WriteString("<a href=\"")
	b.WriteString(t.ContentURL())
	if t.Title != "" {
		b.WriteString("\" title=\"")
		b.WriteString(html.EscapeString(t.Title))
	}
	b.WriteString("\">")
	b.WriteString("view orginal site")
	b.WriteString("</a>")

	b.WriteString("</p>")

	writeFeeds(b, t)

	b.WriteString("</footer>")
}

func writeFeeds(b *strings.Builder, t *pipeline.Task) {
	if len(t.Feeds) == 0 {
		return
	}

	b.WriteString("<p>RSS Feeds:</p>")
	b.WriteString("<ul>")
	for _, fi := range t.Feeds {
		b.WriteString("<li><a href=\"")
		b.WriteString(fi.URL)
		b.WriteString("\">")
		if fi.Title != "" {
			b.WriteString(html.EscapeString(fi.Title))
		} else {
			b.WriteString(fi.URL)
		}
		b.WriteString("</a></li>")
	}
	b.WriteString("</ul")
}
