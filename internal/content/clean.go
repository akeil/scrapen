package content

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

// Clean removes unwanted elements and attributes from the content.
func Clean(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "content",
	}).Info("Clean HTML")

	doc := t.Document()

	resolvePicture(doc)
	normalizeUrls(doc)
	removeUnsupportedSchemes(doc)

	removeUnwantedElements(doc)
	unwrapTags(doc)
	removeUnwantedPunctuation(doc)
	removeUnwantedAttributes(doc)

	dropOrphanedElements(doc)

	stripFromTitle(t)

	return nil
}

// dropUnwantedTags finds tags from the greylists an removes the *markup*
// but keeps the text content
func unwrapTags(doc *goquery.Document) {
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		if isGraylisted(tag) {
			// cannot unwrap empty
			if s.Contents().Length() == 0 {
				s.Remove()
			} else {
				s.Contents().Unwrap()
			}
		}
	})

	unwrapAnchors(doc)
}

func unwrapAnchors(doc *goquery.Document) {
	doc.Selection.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			s.Contents().Unwrap()
		}
	})
}

// remove unwanted attribs
func removeUnwantedAttributes(doc *goquery.Document) {
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		// needs to be done in two steps,
		// editing Node.Attribs is not allowed while iterating
		remove := make([]string, 0)
		for _, attr := range s.Nodes[0].Attr {
			if !isWhitelistedAttr(tag, attr.Key) {
				remove = append(remove, attr.Key)
			}
		}
		for _, key := range remove {
			s.RemoveAttr(key)
		}
	})
}

var punctuation = []string{
	"|",
	"-",
	"/",
	":",
	"#",
	"~",
	"*",
	"+",
}

func isPunctuation(s string) bool {
	s = strings.TrimSpace(s)
	for _, p := range punctuation {
		if s == p {
			return true
		}
	}
	return false
}

// Remove block-level elements that contain only punctuation
// or typical "separators".
func removeUnwantedPunctuation(doc *goquery.Document) {
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		if !isBlocklevel(tag) {
			return
		}
		if isPunctuation(s.Text()) {
			log.WithFields(log.Fields{
				"module":  "content",
				"element": tag,
				"text":    s.Text(),
			}).Debug("Remove element")
			s.Remove()
		}
	})
}

// remove any non-whitelisted tags - including their content/children
func removeUnwantedElements(doc *goquery.Document) {
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		if !isWhitelistedTag(tag) {
			s.Remove()
		}
	})
}

func removeUnsupportedSchemes(doc *goquery.Document) {
	doc.Selection.Find("img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		if src == "" {
			return
		}

		u, err := url.Parse(src)
		if err != nil {
			log.WithFields(log.Fields{
				"module": "content",
				"url":    src,
				"error":  err,
			}).Info("Error parsing image src, removing element")
			s.Remove()
			return
		}

		switch u.Scheme {
		case "data", "http", "https":
			// supported
			return
		case "":
			// empty scheme is supported because we resolve it later
			// against the HTTP(S) base url
			return
		default:
			log.WithFields(log.Fields{
				"module": "content",
				"url":    src,
				"scheme": u.Scheme,
			}).Info("Removing image with unsupported scheme in src")
			s.Remove()
		}
	})
}

func normalizeUrls(doc *goquery.Document) {

	norm := func(raw string) string {
		raw = strings.TrimSpace(raw)

		u, err := url.Parse(raw)
		if err == nil {
			return u.String()
		}

		// in case the src/href *contains* a URL
		found := findURL(raw)
		if found != "" {
			u, err := url.Parse(found)
			if err == nil {
				return u.String()
			}
		}

		return ""
	}

	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href != "" {
			s.SetAttr("href", norm(href))
		}

		src, _ := s.Attr("src")
		if src != "" {
			s.SetAttr("src", norm(src))
		}
	})
}

// regex from daring fireball
// https://daringfireball.net/2010/07/improved_regex_for_matching_urls
// pattern contains a literal Backtick (`) which we need to pseudo-escape with Replace()
const pat = `(?i)\b((?:[a-z][\w-]+:(?:/{1,3}|[a-z0-9%])|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s_BACKTICK_!()\[\]{};:'".,<>?«»“”‘’]))`

var urlPattern = regexp.MustCompile(strings.ReplaceAll(pat, "_BACKTICK_", "´"))

func findURL(s string) string {
	matches := urlPattern.FindAllString(s, 1)
	if len(matches) > 0 {
		return matches[0]
	}

	return ""
}

// find elements that typically require a specific parent element
// and where te parent element has been removed for some reason.
func dropOrphanedElements(doc *goquery.Document) {
	doc.Find("figcaption").Each(func(i int, s *goquery.Selection) {
		found := false
		s.Parents().Each(func(i int, p *goquery.Selection) {
			if goquery.NodeName(p) == "figure" {
				found = true
			}
		})
		if !found {
			s.Remove()
		}
	})
}

// strip prefix or suffix from title
func stripFromTitle(t *pipeline.Task) {
	stripSiteNameFromTitle(t)
	separators := []string{"|"}
	for _, sep := range separators {
		if strings.Count(t.Title, sep) == 1 {
			parts := strings.Split(t.Title, sep)
			a := len(parts[0])
			b := len(parts[1])
			if a > b {
				// assume suffix
				t.Title = strings.TrimSpace(parts[0])
			} else {
				// assume prefix
				t.Title = strings.TrimSpace(parts[1])
			}
		}
	}
}

func stripSiteNameFromTitle(t *pipeline.Task) {
	if t.SiteName == "" {
		return
	}

	siteName := strings.ToLower(t.SiteName)
	title := strings.ToLower(t.Title)
	var prefixLen int
	var suffixLen int
	separators := []string{"|", ":", "-"}
	for _, sep := range separators {
		prefix := siteName + sep
		if strings.HasPrefix(title, prefix) {
			prefixLen = len(prefix)
			break
		}
		prefix = siteName + " " + sep
		if strings.HasPrefix(title, prefix) {
			prefixLen = len(prefix)
			break
		}

		suffix := sep + siteName
		if strings.HasSuffix(title, suffix) {
			suffixLen = len(suffix)
			break
		}
		suffix = sep + " " + siteName
		if strings.HasSuffix(title, suffix) {
			suffixLen = len(suffix)
			break
		}
	}

	if prefixLen > 0 {
		t.Title = strings.TrimSpace(t.Title[prefixLen:])
	} else if suffixLen > 0 {
		n := len(t.Title) - suffixLen
		t.Title = strings.TrimSpace(t.Title[:n])
	}
}
