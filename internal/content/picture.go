package content

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// ConvertAmpImg converts an amp-img to a "normal" img element.
func convertAmpImg(doc *goquery.Document) {
	doc.Selection.Find("amp-img").Each(func(i int, s *goquery.Selection) {
		// Copy all attributes (list of Nodes should be exactly one)
		attrs := s.Nodes[0].Attr

		// Drop children and text content
		s.Contents().Remove()

		// Replace with img element
		node := &html.Node{
			Type:     html.ElementNode,
			Data:     "img",
			DataAtom: atom.Img,
			Attr:     attrs,
		}
		s.ReplaceWithNodes(node)
	})
}

func resolvePicture(doc *goquery.Document) {
	doc.Selection.Find("picture").Each(func(i int, s *goquery.Selection) {
		img := s.Find("img")
		if img.Length() != 1 {
			log.WithFields(log.Fields{
				"module": "content",
			}).Warn("Found multiple img elements in <picture>")

			return
		}

		sources := make([]source, 0)
		s.Find("source").Each(func(j int, source *goquery.Selection) {
			parsed := parseSource(source)
			if len(parsed.srcset) != 0 {
				sources = append(sources, parsed)
			}
		})

		sources = filterSourcesByMediaQuery(sources)

		srcset := selectSrcsetFromSources(sources)
		if srcset.url == "" {
			log.WithFields(log.Fields{
				"module": "content",
			}).Warn("no suitable src found for <picture>")
			return
		}

		img.SetAttr("src", srcset.url)
		s.Find("source").Remove()
		s.Contents().Unwrap()

	})
}

var specialSrc = []string{
	"ix-path", // https://www.imgix.com/
}

// fixSrcs Re-replaces images srcs created by various javascript frameworks
// and establishes a "normal" src value for the image.
func fixSrcs(doc *goquery.Document) {
	// TODO: for imgix, we would also need to look at ix-host="..."
	doc.Selection.Find("img").Each(func(i int, s *goquery.Selection) {
		for _, name := range specialSrc {
			val, _ := s.Attr(name)
			if val != "" {
				log.WithFields(log.Fields{
					"module": "content",
					"src":    val,
				}).Debug("Fixed src")

				s.RemoveAttr("src")
				s.SetAttr("src", val)
			}
		}
	})
}

// ResolveSrcset looks for the srcset attribute in images (img) and selects the
// best (highest resolution) src.
// Replaces the original src.
func resolveSrcset(doc *goquery.Document) {
	doc.Selection.Find("img").Each(func(i int, s *goquery.Selection) {
		srcs, _ := s.Attr("srcset")
		// used by some lazyload JS libs (apparently)
		dataSrcs, _ := s.Attr("data-srcset")
		srcsets := parseSrcset(srcs, dataSrcs)
		set := selectSrcset(srcsets)

		if set.url != "" {
			old, _ := s.Attr("src")
			log.WithFields(log.Fields{
				"module": "content",
				"src":    set.url,
				"old":    old,
			}).Debug("Replaced image src from srcset")

			s.RemoveAttr("src")
			s.SetAttr("src", set.url)
		}
	})
}

func parseSource(s *goquery.Selection) source {
	typ, _ := s.Attr("type")
	media, _ := s.Attr("media")
	srcs, _ := s.Attr("srcset")
	// used by some lazyload JS libs (apparently)
	dataSrcs, _ := s.Attr("data-srcset")

	srcsets := parseSrcset(srcs, dataSrcs)
	mq := parseMediaQueryWidth(media)

	return source{
		contentType: typ,
		media:       mq,
		srcset:      srcsets,
	}
}

func parseSrcset(values ...string) []srcset {
	options := make([]string, 0)
	for _, v := range values {
		parts := strings.Split(v, ",")
		options = append(options, parts...)
	}

	sets := make([]srcset, 0)

	for _, o := range options {
		if strings.TrimSpace(o) == "" {
			continue
		}
		srcset, err := parseSrcsetOption(o)
		if err != nil {
			log.WithFields(log.Fields{
				"module": "content",
				"error":  err,
			}).Warn("Invalid srcset")
		} else {
			sets = append(sets, srcset)
		}
	}

	return sets
}

func parseSrcsetOption(o string) (srcset, error) {
	o = strings.TrimSpace(o)
	parts := strings.Split(o, " ")

	// Empty, e.g. srcset=""
	if len(parts) == 0 {
		return srcset{}, fmt.Errorf("got invalid srcset %q", o)
		// we expect two entries per option (URL and width)
	} else if len(parts) > 2 {
		return srcset{}, fmt.Errorf("invalid srcset: %q", o)
	}

	// The first part is the URL
	set := srcset{url: parts[0]}

	// Optional second part is width and density
	if len(parts) == 2 {
		x := parts[1]

		if strings.HasSuffix(x, "w") {
			width, err := strconv.ParseUint(strings.TrimSuffix(x, "w"), 10, 64)
			if err != nil {
				return srcset{}, fmt.Errorf("failed to parse int from %q", x)
			}
			set.width = width

		} else if strings.HasSuffix(x, "x") {
			density, err := strconv.ParseFloat(strings.TrimSuffix(x, "x"), 64)
			if err != nil {
				return srcset{}, fmt.Errorf("failed to parse float from %q", x)
			}
			set.density = density
		}
	}

	return set, nil
}

var widthMediaQuery = regexp.MustCompile(`.*?\(((?:min-width|max-width|width):\s*[0-9]+)px\).*?`)

// Can only handle a single mediaquery for min-width, max-width or width
// https://developer.mozilla.org/en-US/docs/Web/CSS/Media_Queries/Using_media_queries#media_features
func parseMediaQueryWidth(s string) mediaQuery {
	result := mediaQuery{}
	// should give us exactly two matches,
	// one for the whole pattern, another for the capturing group
	matches := widthMediaQuery.FindStringSubmatch(s)
	if len(matches) != 2 {
		return result
	}

	// matche should look like this:
	// max-width:  1024px
	m := matches[1]

	parts := strings.Split(m, ":")
	if len(parts) != 2 {
		return result
	}

	name := strings.TrimSpace(parts[0])
	val, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"raw":        parts[1],
			"mediaQuery": s,
		}).Warning("Failed to parse integer from media query")
		return result
	}

	switch name {
	case "min-width":
		result.minWidth = val
		return result
	case "width":
		result.width = val
		return result
	case "max-width":
		result.maxWidth = val
		return result
	}

	return result
}

func filterSourcesByMediaQuery(sources []source) []source {
	var selected source

	// by media query width
	maxWidth := 0.0
	for _, s := range sources {
		if s.media.IsEmpty() {
			continue
		}
		w := math.Max(float64(s.media.minWidth), float64(s.media.maxWidth))
		w = math.Max(w, float64(s.media.width))
		if w > maxWidth {
			maxWidth = w
			selected = s
		}
	}

	if !selected.media.IsEmpty() {
		return []source{selected}
	}

	return sources
}

func selectSrcsetFromSources(sources []source) srcset {
	sets := make([]srcset, 0)
	for _, source := range sources {
		sets = append(sets, source.srcset...)
	}

	return selectSrcset(sets)
}

func selectSrcset(sets []srcset) srcset {
	if len(sets) == 0 {
		return srcset{}
	}

	var maxWidth uint64
	var maxDensity float64
	var byWidth srcset
	var byDensity srcset

	for _, s := range sets {
		if s.width != 0 && s.width > maxWidth {
			maxWidth = s.width
			byWidth = s
		} else if s.density != 0 && s.density > maxDensity {
			maxDensity = s.density
			byDensity = s
		}
	}

	if maxWidth != 0 {
		return byWidth
	} else if maxDensity != 0 {
		return byDensity
	}

	// fallback: return the first entry
	return sets[0]
}

type source struct {
	contentType string
	media       mediaQuery
	srcset      []srcset
}

type srcset struct {
	url     string
	width   uint64
	density float64
}

type mediaQuery struct {
	minWidth int
	width    int
	maxWidth int
}

func (m mediaQuery) IsEmpty() bool {
	return m.minWidth == 0 && m.maxWidth == 0 && m.width == 0
}
