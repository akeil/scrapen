package content

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

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

		srcset := selectSrcset(sources)
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

func parseSource(s *goquery.Selection) source {
	typ, _ := s.Attr("type")
	media, _ := s.Attr("media")
	srcs, _ := s.Attr("srcset")
	// used by some lazyload JS libs (apparently)
	dataSrcs, _ := s.Attr("data-srcset")

	srcsets := make([]srcset, 0)
	options := strings.Split(srcs, ",")
	dataOpts := strings.Split(dataSrcs, ",")
	options = append(options, dataOpts...)

	for _, o := range options {
		o = strings.TrimSpace(o)
		parts := strings.Split(o, " ")
		if len(parts) == 0 {
			return source{}
		} else if len(parts) > 2 {
			log.WithFields(log.Fields{
				"module": "content",
			}).Warn(fmt.Sprintf("Invalid srcset: %q", o))

			return source{}
		}

		url := parts[0]
		srcset := srcset{url: url}

		if len(parts) == 2 {
			x := parts[1]

			if strings.HasSuffix(x, "w") {
				width, err := strconv.ParseUint(strings.TrimSuffix(x, "w"), 10, 64)
				if err != nil {
					log.WithFields(log.Fields{
						"module": "content",
					}).Warn(fmt.Sprintf("Failed to parse int from %q", x))

					return source{}
				}
				srcset.width = width

			} else if strings.HasSuffix(x, "x") {
				density, err := strconv.ParseFloat(strings.TrimSuffix(x, "x"), 64)
				if err != nil {
					log.WithFields(log.Fields{
						"module": "content",
					}).Warn(fmt.Sprintf("Failed to parse float from %q", x))

					return source{}
				}
				srcset.density = density
			}

		}
		srcsets = append(srcsets, srcset)
	}

	return source{
		contentType: typ,
		media:       media,
		srcset:      srcsets,
	}
}

func selectSrcset(sources []source) srcset {
	if len(sources) == 0 {
		return srcset{}
	}

	var maxWidth uint64
	var maxDensity float64
	var byWidth srcset
	var byDensity srcset

	for _, source := range sources {
		for _, srcset := range source.srcset {
			if srcset.width != 0 && srcset.width > maxWidth {
				maxWidth = srcset.width
				byWidth = srcset
			} else if srcset.density != 0 && srcset.density > maxDensity {
				maxDensity = srcset.density
				byDensity = srcset
			}
		}
	}

	if maxWidth != 0 {
		return byWidth
	} else if maxDensity != 0 {
		return byDensity
	}

	// falback, simply return the last <source> entry, first url
	s := sources[len(sources)-1].srcset
	return s[0]
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
