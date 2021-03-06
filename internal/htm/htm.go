package htm

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/akeil/scrapen/internal/pipeline"
)

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
		b.WriteString(t.PubDate.Format(time.ANSIC))
		b.WriteString("</time>")
		b.WriteString("</p>")
	}

	if t.Site != "" {
		b.WriteString("<p>")
		b.WriteString(t.Site)
		b.WriteString("</p>")
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
	b.WriteString(t.Retrieved.Format(time.ANSIC))
	b.WriteString("</time>")
	b.WriteString(" | ")

	b.WriteString("<a href=\"")
	b.WriteString(t.ContentURL())
	if t.Title != "" {
		b.WriteString("\" title=\"")
		// TODO: Escape?
		b.WriteString(t.Title)
	}
	b.WriteString("\">")
	b.WriteString("view orginal site")
	b.WriteString("</a>")

	b.WriteString("</p>")
	b.WriteString("</footer>")
}

// TODO: read this from a file or embed
const style = `body {
	background: #ffffff;
	font-family: sans-serif;
	margin: 3em;
}

h1, h2, h3, h4, h5, h6 {
	font-family: serif;
}

a {
	color: #007bff; /* light blue */
	text-decoration: none;
}

dl {
	display: block;
	margin-top: 0;
	margin-bottom: 1em;
	border-left: 1px solid #cccccc;
	padding-left: 0.25em;
}

dt {
	display: block;
	clear: left;
	float: left;
	margin: 0;
	padding: 0 0.5em 0 0;
	font-weight: bold;
}

dd {
	display: block;
	margin: 0 0 0.5em 2em;
}

code {
	color: #e83e8c; /* pink */
	font-family: monospace;
}

pre {
	font-family: monospace;
	white-space: pre-wrap;
	line-height: 125%;
	background: #f8f8f8;
	border: 1px solid #cccccc;
	border-radius: 0.25em;
	margin-left: 1em;
	margin-right: 1em;
	margin-bottom: 1em;
	margin-bottom: 0;
	padding: 0.75em;
}

figcaption {
	font-style: italic;
	font-size: smaller;
}

time {
	font-style: italic;
}

footer {
	border-top: 1px solid #cccccc;
	font-size: smaller;
}

header {
	border-bottom: 1px solid #cccccc;
	color: #909090;
}

header p {
	font-weight: normal;
	font-size: small;
}

`
