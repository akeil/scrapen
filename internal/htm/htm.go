package htm

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"akeil.net/akeil/elsewhere/internal/pipeline"
)

func Compose(w io.Writer, i *pipeline.Item) error {
	var b strings.Builder

	b.WriteString("<html>")
	writeHead(&b, i)
	err := writeBody(&b, i)
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

func writeHead(b *strings.Builder, i *pipeline.Item) {
	b.WriteString("<head>")
	b.WriteString("<meta charset=\"utf-8\"/>")
	b.WriteString(fmt.Sprintf("<title>%v</title>", html.EscapeString(i.Title)))
	b.WriteString("<style>")
	b.WriteString(style)
	b.WriteString("</style>")
	b.WriteString("</head>")
}

func writeBody(b *strings.Builder, i *pipeline.Item) error {
	b.WriteString("<body>")
	err := writeContent(b, i)
	if err != nil {
		return err
	}
	writeFooter(b, i)
	b.WriteString("</body>")
	return nil
}

func writeContent(b *strings.Builder, i *pipeline.Item) error {
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
			err = dataImage(t.Attr, i, w)
			if err != nil {
				return false, err
			}
			w.WriteString(">")
			return true, nil
		case html.SelfClosingTagToken:
			w.WriteString("<")
			w.WriteString(t.Data)
			err = dataImage(t.Attr, i, w)
			if err != nil {
				return false, err
			}
			w.WriteString("/>")
			return true, nil
		}

		return false, nil
	}

	return pipeline.WalkHTML(b, i.Html, handler)
}

func dataImage(a []html.Attribute, i *pipeline.Item, w io.StringWriter) error {
	for _, attr := range a {
		if attr.Key == "src" {
			u, err := url.Parse(attr.Val)
			if err != nil {
				return err
			}
			if u.Scheme == "local" {
				contentType, data, err := i.GetAsset(u.Host)
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

func writeFooter(b *strings.Builder, i *pipeline.Item) {
	b.WriteString("<footer>")
	b.WriteString("<p>")

	b.WriteString("Retrieved on ")
	// see:
	// http://microformats.org/wiki/datetime-design-pattern
	b.WriteString("<time datetime=\"")
	b.WriteString(i.Retrieved.Format(time.RFC3339))
	b.WriteString("\">")
	b.WriteString(i.Retrieved.Format(time.ANSIC))
	b.WriteString("</time>")
	b.WriteString(" | ")

	b.WriteString("<a href=\"")
	b.WriteString(i.Url)
	if i.Title != "" {
		b.WriteString("\" title=\"")
		// TODO: Escape?
		b.WriteString(i.Title)
	}
	b.WriteString("\">")
	b.WriteString("view orginal site")
	b.WriteString("</a>")

	b.WriteString("</p>")
	b.WriteString("</footer>")
}

// TODO: read this from a file
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
`
