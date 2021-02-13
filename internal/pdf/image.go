package pdf

import (
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

type image struct {
	pdf gofpdf.Pdf
	src string
}

func newImage(pdf gofpdf.Pdf, attr map[string]string) renderer {
	return &image{
		pdf: pdf,
		src: attr["src"],
	}
}

func (i *image) Text(s string) {}
func (i *image) End()          {}
func (i *image) Tag(tag string, attr map[string]string) renderer {
	fmt.Printf("Tag within image: %q\n", tag)
	return FindRenderer(tag, attr, i.pdf)
}

func (i *image) EndTag(s string) {
	fmt.Printf("End image with src %q\n", i.src)
	i.addImage()
}

func (i *image) addImage() {
	if i.src == "" {
		return
	}

	name := i.src
	x := 0.0
	y := 0.0
	w := 0.0
	h := 0.0
	opts := gofpdf.ImageOptions{
		ImageType:             "JPEG",
		ReadDpi:               true,
		AllowNegativePosition: false,
	}
	flow := true // advances y-pos
	linkID := 0
	linkText := ""
	i.pdf.ImageOptions(name, x, y, w, h, flow, opts, linkID, linkText)
}
