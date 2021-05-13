package metadata

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

// FallbackImage sets the main image for an article from a fallback source
// if no other image has been set.
func FallbackImage(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "metadata",
		"url":    t.ContentURL(),
	}).Info("Fallback Image")

	if t.ImageURL != "" {
		return nil
	}

	fallbackImagefromContent(t)
	fallbackImageFromIcon(t)

	return nil
}

// set the article image from the first image we find in the content.
func fallbackImagefromContent(t *pipeline.Task) {
	doc := t.Document()
	doc.Selection.Find("img").First().Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		t.ImageURL = src
	})
}

var iconRels = []string{
	"icon",
	"apple-touch-icon",
	"mask-icon",
	"shortcut icon",
}

func fallbackImageFromIcon(t *pipeline.Task) {
	doc := t.Document()
	var icons iconList
	icons = make(iconList, 0)
	// link rel="icon" w/ multiple sizes="16x16"
	// link rel="mask-icon"
	// link rel=shortcut icon"
	// link rel=apple-touch-icon
	doc.Selection.Find("link").Each(func(i int, s *goquery.Selection) {
		rel, _ := s.Attr("rel")
		if !contains(iconRels, rel) {
			return
		}

		href, _ := s.Attr("href")
		if href == "" {
			return
		}

		sizes, _ := s.Attr("sizes")
		icons = append(icons, icon{rel, href, sizes})
	})

	// meta name="msapplication-TileImage content=""
	// link rel=manifest
	// ... and icons = [...]

	sort.Sort(icons)
	if len(icons) > 0 {
		t.ImageURL = icons[0].href
	}
}

type icon struct {
	name string
	href string
	size string
}

func (i icon) area() int {
	if i.size == "" {
		return 0
	}
	parts := strings.Split(strings.ToLower(i.size), "x")
	if len(parts) != 2 {
		return 0
	}

	w, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}

	h, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}

	return w * h
}

// sort interface

type iconList []icon

func (l iconList) Len() int {
	return len(l)
}

func (l iconList) Less(i, j int) bool {
	// Tell if "a" is before "b"
	a := l[i]
	b := l[j]

	// compare sizes
	aArea := a.area()
	bArea := b.area()
	if aArea != 0 && bArea != 0 {
		return aArea > bArea
	}

	// by preference
	aPref := len(iconRels)
	bPref := len(iconRels)
	for idx, val := range iconRels {
		if val == a.name {
			aPref = idx
		}
		if val == b.name {
			bPref = idx
		}
	}
	if aPref != bPref {
		return aPref < bPref
	}

	// if no other sort criteria is found, leave unchanged
	return i < j
}

func (l iconList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
