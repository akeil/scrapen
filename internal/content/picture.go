package content

import (
	"fmt"
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
			log.WithFields(log.Fields{
				"module": "content",
				"src":    set.url,
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

	return source{
		contentType: typ,
		media:       media,
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
		return srcset{}, fmt.Errorf("Invalid srcset: %q", o)
	}

	// The first part is the URL
	set := srcset{url: parts[0]}

	// Optional second part is width and density
	if len(parts) == 2 {
		x := parts[1]

		if strings.HasSuffix(x, "w") {
			width, err := strconv.ParseUint(strings.TrimSuffix(x, "w"), 10, 64)
			if err != nil {
				return srcset{}, fmt.Errorf("Failed to parse int from %q", x)
			}
			set.width = width

		} else if strings.HasSuffix(x, "x") {
			density, err := strconv.ParseFloat(strings.TrimSuffix(x, "x"), 64)
			if err != nil {
				return srcset{}, fmt.Errorf("Failed to parse float from %q", x)
			}
			set.density = density
		}
	}

	return set, nil
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
	media       string
	srcset      []srcset
}

type srcset struct {
	url     string
	width   uint64
	density float64
}
