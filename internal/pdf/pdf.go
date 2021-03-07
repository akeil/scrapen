package pdf

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/jung-kurt/gofpdf"
	"golang.org/x/net/html"

	"github.com/akeil/scrapen/internal/pipeline"
)

func Compose(w io.Writer, t *pipeline.Task) error {
	b := newBuilder(t)
	err := b.build()
	if err != nil {
		return err
	}

	return b.doc.Output(w)
}

type builder struct {
	task *pipeline.Task
	doc  gofpdf.Pdf
	dom  *goquery.Document
}

func newBuilder(t *pipeline.Task) *builder {
	b := &builder{
		task: t,
	}
	return b
}

func (b *builder) build() error {
	err := b.parse()
	if err != nil {
		return err
	}
	b.setup()
	b.doc.AddPage()
	b.title()
	err = b.walk()
	if err != nil {
		return err
	}

	return nil
}

func (b *builder) title() {
	b.doc.SetFont("Times", "B", 24.0)
	b.doc.Cell(60, 14.0, b.task.Title)
	b.doc.Ln(14) // line break
}

func (b *builder) walk() error {
	r := strings.NewReader(b.task.HTML)
	z := html.NewTokenizer(r)

	var c renderer
	c = &nullRenderer{b.doc}

	for {
		t := z.Next()
		fmt.Printf("Token: %v\n", t)
		switch t {
		case html.ErrorToken:
			err := z.Err()
			if err == io.EOF {
				return nil
			}
			return z.Err()

		case html.TextToken:
			text := string(z.Text())
			fmt.Printf("T: %q\n", text)
			c.Text(text)

		case html.StartTagToken:
			rawTag, hasAttr := z.TagName()
			tag := string(rawTag)
			var attribs map[string]string
			if hasAttr {
				attribs = make(map[string]string)
				for {
					k, v, more := z.TagAttr()
					if !more {
						break
					}
					attribs[string(k)] = string(v)
				}
			}
			fmt.Printf(" + %v\n", tag)
			c = c.Tag(tag, attribs)

		case html.EndTagToken:
			rawTag, _ := z.TagName()
			tag := string(rawTag)
			fmt.Printf(" - %v\n", tag)
			c.EndTag(tag)

		case html.SelfClosingTagToken:
			rawTag, hasAttr := z.TagName()
			tag := string(rawTag)
			var attribs map[string]string
			if hasAttr {
				attribs = make(map[string]string)
				for {
					k, v, more := z.TagAttr()
					if !more {
						break
					}
					attribs[string(k)] = string(v)
				}
			}
			fmt.Printf(" + %v\n", tag)
			c = c.Tag(tag, attribs)
			c.EndTag(tag)
		}

	}
}

func (b *builder) setup() {
	orientation := "P" // [P]orttrait
	unit := "mm"
	size := "A4"
	fontDir := "./fonts"
	b.doc = gofpdf.New(orientation, unit, size, fontDir)

	// sets the default font
	// Courier, Helvetica, Times
	// B=bold, I=italic, U=underscore, S=strikeout + combos (e.g. "IB")
	b.doc.SetFont("Helvetica", "", 10.0)
}

func (b *builder) parse() error {
	r := strings.NewReader(b.task.HTML)
	dom, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}
	b.dom = dom
	return nil
}

type renderer interface {
	Text(s string)
	End()
	Tag(tag string, attr map[string]string) renderer
	EndTag(tag string)
}

func FindRenderer(tag string, attr map[string]string, pdf gofpdf.Pdf) renderer {
	fmt.Printf("Find renderer for %q\n", tag)

	switch tag {
	case "p":
		return newParagraph(pdf, attr, "Helvetica", 12.0)
	case "figcaption":
		return newParagraph(pdf, attr, "Helvetica", 10.0)
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level, err := strconv.Atoi(tag[1:])
		if err != nil {
			// this really should not happen
			return &nullRenderer{pdf}
		}
		return newHeading(pdf, attr, level)
	case "img", "figure":
		return newImage(pdf, attr)
	default:
		return &nullRenderer{pdf}
	}
}

type nullRenderer struct {
	pdf gofpdf.Pdf
}

func (n *nullRenderer) Text(s string) {}
func (n *nullRenderer) End()          {}
func (n *nullRenderer) Tag(tag string, attr map[string]string) renderer {
	return FindRenderer(tag, attr, n.pdf)
}
func (n *nullRenderer) EndTag(s string) {}

func collapseWhitespace(s string) string {
	prefix := ""
	suffix := ""
	if unicode.IsSpace(rune(s[0])) {
		prefix = " "
	}
	if len(s) > 2 && unicode.IsSpace(rune(s[len(s)-1])) {
		suffix = " "
	}

	parts := strings.Fields(s)
	return prefix + strings.Join(parts, " ") + suffix
}
