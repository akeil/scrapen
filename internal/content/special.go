package content

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/akeil/scrapen/internal/pipeline"
)

func resolveIFrames(doc *goquery.Document) {
	log.Debug("looking for iframe...")
	doc.Selection.Find("iframe").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")

		log.WithFields(log.Fields{
			"src":    src,
			"module": "content",
		}).Debug("Found iframe")

		u, err := url.Parse(src)
		if err != nil {
			log.WithFields(log.Fields{
				"src":    src,
				"module": "content",
			}).Warning("Failed to parse iframe src")
			return
		}

		handleIFrame(u, s)
	})
}

func handleIFrame(src *url.URL, s *goquery.Selection) {
	h := strings.TrimPrefix(src.Host, "www.")
	log.Debug(h)
	switch h {
	case "youtube.com":
		youtubeVideo(src, s)
	}
}

func youtubeVideo(src *url.URL, s *goquery.Selection) {
	// https://www.youtube.com/embed/lC8T4HXrkpk?feature=oembed
	log.Debug("found youtube iframe")

	parts := strings.Split(src.Path, "/")
	// the path is absolute and the first component is empty
	if parts[1] != "embed" {
		return
	}

	videoID := parts[2]

	// TODO: we could retrieve the video title and display it
	// TODO: we might add the video URL to a list of media attachments

	// Replace with img element
	href := fmt.Sprintf("https://youtube.com/watch?v=%v", videoID)
	attrs := []html.Attribute{
		html.Attribute{Key: "href", Val: href},
	}
	a := &html.Node{
		Type:     html.ElementNode,
		Data:     "a",
		DataAtom: atom.Img,
		Attr:     attrs,
	}

	title, _ := s.Attr("title")
	var text string
	if title == "" {
		text = "Watch this video on YouTube."
	} else {
		text = fmt.Sprintf("Watch the video %q on YouTube.", title)
	}
	a.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: text,
	})

	cap := &html.Node{
		Type:     html.ElementNode,
		Data:     "figcaption",
		DataAtom: atom.Figcaption,
	}
	cap.AppendChild(a)

	// thumbnail image:
	// - https://img.youtube.com/vi/VIDEO-ID/QUALITY.jpg
	//
	// Quality: default, hqdefault, mqdefault, sddefault, maxresdefault
	img := &html.Node{
		Type:     html.ElementNode,
		Data:     "img",
		DataAtom: atom.Img,
		Attr: []html.Attribute{
			html.Attribute{
				Key: "src",
				Val: fmt.Sprintf("https://img.youtube.com/vi/%v/%v.jpg", videoID, "mqdefault"),
			},
			html.Attribute{
				Key: "width",
				Val: "400",
			},
		},
	}

	fig := &html.Node{
		Type:     html.ElementNode,
		Data:     "figure",
		DataAtom: atom.Figure,
	}

	fig.AppendChild(img)
	fig.AppendChild(cap)

	s.ReplaceWithNodes(fig)
}

// jsonLD handles JSON-LD markup as described on https://schema.org/
// see: https://moz.com/blog/json-ld-for-beginners
func jsonLD(t *pipeline.Task) {
	doc := t.Document()
	// looking for   <script type="application/ld+json">...</script>
	doc.Selection.Find("script").Each(func(i int, s *goquery.Selection) {
		tp, _ := s.Attr("type")
		if tp != "application/ld+json" {
			return
		}

		data := make(map[string]interface{})
		err := json.Unmarshal([]byte(s.Text()), &data)
		if err != nil {
			log.WithFields(log.Fields{
				"module": "content",
				"error":  err,
			}).Warning("Failed t parse JSON-LD")
			return
		}

		if ldType, ok := data["@type"].(string); ok {
			switch ldType {
			case "Audio":
				ldAudio(t, data)
				break
			case "Article":
				ldArticle(t, data)
				break
			case "Organization":
				break
			case "BreadcrumbList":
				break
			}
		}
	})
}

func ldAudio(t *pipeline.Task, m map[string]interface{}) {
	/* {
	   "@context": "http://schema.org",
	   "@type": "Audio",
	   "name": "Migration - Das US-Einwanderungsgesetz von 1921",
	   "description": "",
	   "contentUrl": "https://ondemand-mp3.dradio.de/file/dradio/2021/05/17/deutschlandfunknova_migration_das_20210517_8a14882a.mp3",
	   "encodingFormat": "audio/mpeg",
	   "contentSize": "39285987",
	   "transcript": "",
	   "uploadDate": "2021-05-04",
	   "duration": "PT41M1S",
	   "inLanguage": {
	       "@type": "Language",
	       "name": "German",
	       "alternateName": "de"
	   },
	   "productionCompany": {
	       "@type": "Organization",
	       "name": "Deutschlandfunk Nova"
	   }
	   } */
	if u, ok := m["contentUrl"].(string); ok {
		u, err := t.ResolveURL(u)
		if err != nil {
			return
		}

		name, _ := m["name"].(string)
		ct, _ := m["encodingFormat"].(string)
		desc, _ := m["description"].(string)

		enc := pipeline.Enclosure{
			Type:        "Audio",
			Title:       name,
			URL:         u,
			ContentType: ct,
			Description: desc,
		}
		log.Info("Add audio enclosure")
		t.AddEnclosure(enc)
	}
}

func ldArticle(t *pipeline.Task, m map[string]interface{}) {
	/*  {
	    "@context": "http://schema.org",
	    "@type": "Article",
	    "headline": "Der Emergency Quota Act von 1921",
	    "image": ["https://static.deutschlandfunknova.de/editorial/_entryImage/210514_ellis_island_thumb.jpg?mtime=20210514092308&amp;focal=none&amp;tmtime=20210514123335"],
	    "author":
	    {
	        "@type": "Organization",
	        "name": "Deutschlandfunk Nova"
	    },
	    "publisher": {
	        "@type": "Organization",
	        "name": "Deutschlandfunk Nova",
	        "logo": {
	            "@type": "ImageObject",
	            "url": "https://www.deutschlandfunknova.de/img/dlfnova-g.png"
	        }
	    },
	    "datePublished": "2021-05-14",
	    "dateModified": "2021-05-14",
	    "description": "Der Emergency Quota Act von 1921 schloss alle Menschen aus asiatischen Ländern und Osteuropa von der Einwanderung in die USA aus.",
	    "mainEntityOfPage": "https://www.deutschlandfunknova.de/beitrag/emergency-quota-act-einwanderung-in-die-usa-ab-1921"
	    }     */
}
