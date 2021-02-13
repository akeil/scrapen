package pipeline

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

type TokenHandler func(html.Token, io.StringWriter) (bool, error)

func WalkHTML(w io.StringWriter, s string, h TokenHandler) error {
	var err error
	z := html.NewTokenizer(strings.NewReader(s))

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			err = z.Err()
			if err == io.EOF {
				// Done and OK
				return nil
			}
			return err
		}

		t := z.Token()
		consumed, err := h(t, w)
		if err != nil {
			return err
		}
		if !consumed {
			identity(t, w)
		}
	}
}

func identity(t html.Token, w io.StringWriter) {
	tt := t.Type
	switch tt {
	case html.TextToken:
		w.WriteString(html.EscapeString(t.Data))
	case html.DoctypeToken,
		html.CommentToken:
		// ignored
	case html.StartTagToken:
		w.WriteString("<")
		w.WriteString(t.Data)
		WriteAttrList(t.Attr, w)
		w.WriteString(">")
	case html.EndTagToken:
		w.WriteString("</")
		w.WriteString(t.Data)
		w.WriteString(">")
	case html.SelfClosingTagToken:
		w.WriteString("<")
		w.WriteString(t.Data)
		WriteAttrList(t.Attr, w)
		w.WriteString("/>")
	}
}

func WriteAttrList(a []html.Attribute, w io.StringWriter) {
	for _, attr := range a {
		WriteAttr(attr, w)
	}
}

func WriteAttr(a html.Attribute, w io.StringWriter) {
	w.WriteString(" ")
	w.WriteString(a.Key)
	w.WriteString("=\"")
	w.WriteString(a.Val) // TODO: escape?
	w.WriteString("\"")
}
