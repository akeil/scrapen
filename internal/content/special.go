package content

import (
    "net/url"
    "strings"
    "fmt"

    "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)


func resolveIFrames(doc *goquery.Document) {
    log.Debug("looking for iframe...")
    doc.Selection.Find("iframe").Each(func(i int, s *goquery.Selection) {
        src, _ := s.Attr("src")
        title, _ := s.Attr("title")

        log.WithFields(log.Fields{
    		"src": src,
            "title": title,
    		"module": "content",
    	}).Debug("Found iframe")

        u, err := url.Parse(src)
        if err != nil {
            log.WithFields(log.Fields{
        		"src": src,
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
    a.AppendChild(&html.Node{
        Type: html.TextNode,
        Data: "Watch this video on youtube.",
    })

    cap := &html.Node{
        Type: html.ElementNode,
        Data: "figcaption",
        DataAtom: atom.Figcaption,
    }
    cap.AppendChild(a)

    // thumbnail image:
    // - https://img.youtube.com/vi/VIDEO-ID/QUALITY.jpg
    //
    // Quality: default, hqdefault, mqdefault, sddefault, maxresdefault
    img := &html.Node{
        Type: html.ElementNode,
        Data: "img",
        DataAtom: atom.Img,
        Attr: []html.Attribute{
            html.Attribute{
                Key: "src",
                Val: fmt.Sprintf("https://img.youtube.com/vi/%v/%v.jpg", videoID, "mqdefault"),
            },
            html.Attribute{
                Key: "width",
                Val: "200",
            },
        },
    }

    fig := &html.Node{
        Type: html.ElementNode,
        Data: "figure",
        DataAtom: atom.Figure,
    }

    fig.AppendChild(img)
    fig.AppendChild(cap)

    s.ReplaceWithNodes(fig)
}
