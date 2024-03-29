package pipeline

import (
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type TokenReader func(html.Token) error

func ReadHTML(s string, r TokenReader) error {
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
		err := r(t)
		if err != nil {
			return err
		}

		// working around the HTML parser not parsing the content of <noscript>
		// https://github.com/golang/go/issues/16318
		if t.DataAtom == atom.Noscript {
			z.Next()
			t = z.Token() // the text content of <noscript>
			err = ReadHTML(t.Data, r)
			if err != nil {
				return err
			}
		}
	}
}

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

		// working around the HTML parser not parsing the content of <noscript>
		// https://github.com/golang/go/issues/16318
		if t.DataAtom == atom.Noscript {
			z.Next()
			t = z.Token() // the text content of <noscript>
			err = WalkHTML(w, t.Data, h)
			if err != nil {
				return err
			}
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

// ReadAttr reads the value of the first attribute with the given name.
func ReadAttr(t html.Token, name string) (bool, string) {
	for _, a := range t.Attr {
		if a.Key == name {
			return true, a.Val
		}
	}

	return false, ""
}
