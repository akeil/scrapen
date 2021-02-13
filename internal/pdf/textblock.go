package pdf

import (
	"github.com/jung-kurt/gofpdf"
)

type paragraph struct {
	pdf        gofpdf.Pdf
	fontFamily string
	fontStyle  int
	fontSize   float64
	lineHeight float64
}

func newParagraph(pdf gofpdf.Pdf, attr map[string]string, fontFamily string, fontSize float64) *paragraph {
	return &paragraph{
		pdf:        pdf,
		fontFamily: fontFamily,
		fontStyle:  regular,
		fontSize:   fontSize,
		lineHeight: 1.75,
	}
}

func (p *paragraph) Text(s string) {
	if s == "" {
		return
	}

	p.pdf.SetFont(p.fontFamily, fontStyle(p.fontStyle), p.fontSize)

	s = collapseWhitespace(s)

	unit := p.pdf.PointToUnitConvert(p.fontSize)
	h := unit * p.lineHeight
	p.pdf.Write(h, s)
}

func (p *paragraph) End() {
	unit := p.pdf.PointToUnitConvert(p.fontSize)
	p.pdf.Ln(unit * 3)
}

func (p *paragraph) Tag(tag string, attr map[string]string) renderer {
	// if inline, change font etc.
	switch tag {
	case "br", "hr":
		unit := p.pdf.PointToUnitConvert(p.fontSize)
		h := unit * p.lineHeight
		p.pdf.Write(h, "\n")
		return p

	case "strong", "b":
		p.fontStyle = p.fontStyle | bold
		return p
	case "em", "i":
		p.fontStyle = p.fontStyle | italic
		return p
	case "a":
		p.fontStyle = p.fontStyle | underline
		// TODO: inline links
		return p

	default:
		p.End()
		return FindRenderer(tag, attr, p.pdf)
	}
}

func (p *paragraph) EndTag(tag string) {
	// if inline, reset font
	// move back to parent, end current.
	// make sure we End() only once
	switch tag {
	case "strong", "b":
		p.fontStyle = p.fontStyle &^ bold
	case "em", "i":
		p.fontStyle = p.fontStyle &^ italic
	case "a":
		p.fontStyle = p.fontStyle &^ underline
		// TODO: inline links
	}
}

var fontSize = map[int]float64{
	1: 20.0,
	2: 18.0,
	3: 16.0,
	4: 14.0,
	5: 12.0,
	6: 10.0,
}

func newHeading(pdf gofpdf.Pdf, attr map[string]string, level int) renderer {
	fs, ok := fontSize[level]
	if !ok {
		fs = 10.0
	}
	block := newParagraph(pdf, attr, "Times", fs)
	block.fontStyle = bold
	return block
}
