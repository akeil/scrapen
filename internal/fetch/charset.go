package fetch

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"

	"github.com/akeil/scrapen/internal/pipeline"
)

// readUTF8 reads the response body into a UTF-8 string
func readUTF8(t *pipeline.Task, r io.Reader, h http.Header) (string, error) {
	cs := charsetFromHeader(t, h)

	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "fetch",
	}).Info(fmt.Sprintf("Got charset %q from header", cs))

	// we nned to copy the reader into a buffer so we can read it multiple times.
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, r)
	if err != nil {
		return "", err
	}

	if cs == "" {
		cs = charsetFromMeta(bytes.NewBuffer(buf.Bytes()))
		log.WithFields(log.Fields{
			"task":   t.ID,
			"module": "fetch",
		}).Info(fmt.Sprintf("Got charset %q from meta tag", cs))
	}

	// TODO: we can also obtain the charset from XML declaration

	// needs type io.Reader to wrap into decoder
	var rdr io.Reader
	rdr = buf

	if normalizeCharsetName(cs) != "utf-8" && cs != "" {
		dec := decoderByName(cs)
		if dec != nil {
			rdr = dec.Reader(rdr)

			log.WithFields(log.Fields{
				"task":   t.ID,
				"module": "fetch",
			}).Info(fmt.Sprintf("Found decoder for charset %q", cs))

		} else {
			log.WithFields(log.Fields{
				"task":   t.ID,
				"module": "fetch",
			}).Warn(fmt.Sprintf("Could not find decoder for charset %q, assume UTF-8", cs))
		}
	}

	data, err := io.ReadAll(rdr)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func charsetFromHeader(t *pipeline.Task, h http.Header) string {
	contentType, ok := h["Content-Type"]
	if !ok {
		return ""
	}

	for _, s := range contentType {
		_, params, err := mime.ParseMediaType(s)
		if err != nil {
			log.WithFields(log.Fields{
				"task":   t.ID,
				"module": "fetch",
				"error":  err,
			}).Warn("Error parsing media type")

			return ""
		}
		cs := params["charset"]
		if cs != "" {
			return cs
		}
	}

	return ""
}

func charsetFromMeta(r io.Reader) string {
	var charset string

	reader := func(t html.Token) error {
		if charset != "" {
			// returns error to exit early, error is later ignored
			return fmt.Errorf("charset already set")
		}

		if t.DataAtom == atom.Meta {
			for _, attr := range t.Attr {
				if attr.Key == "charset" {
					charset = attr.Val
				}
			}
		}
		return nil
	}

	var b strings.Builder
	_, err := io.Copy(&b, r)
	if err != nil {
		return ""
	}
	pipeline.ReadHTML(b.String(), reader)

	// set in the reader function
	return charset
}

var charmaps = []*charmap.Charmap{
	charmap.CodePage037,
	charmap.CodePage1047,
	charmap.CodePage1140,
	charmap.CodePage437,
	charmap.CodePage850,
	charmap.CodePage852,
	charmap.CodePage855,
	charmap.CodePage858,
	charmap.CodePage860,
	charmap.CodePage862,
	charmap.CodePage863,
	charmap.CodePage865,
	charmap.CodePage866,

	charmap.ISO8859_1,
	charmap.ISO8859_10,
	charmap.ISO8859_13,
	charmap.ISO8859_14,
	charmap.ISO8859_15,
	charmap.ISO8859_16,
	charmap.ISO8859_2,
	charmap.ISO8859_3,
	charmap.ISO8859_4,
	charmap.ISO8859_5,
	charmap.ISO8859_6,
	charmap.ISO8859_7,
	charmap.ISO8859_8,
	charmap.ISO8859_9,

	charmap.KOI8R,
	charmap.KOI8U,

	charmap.Macintosh,
	charmap.MacintoshCyrillic,

	charmap.Windows1250,
	charmap.Windows1251,
	charmap.Windows1252,
	charmap.Windows1253,
	charmap.Windows1254,
	charmap.Windows1255,
	charmap.Windows1256,
	charmap.Windows1257,
	charmap.Windows1258,
	charmap.Windows874,
}

func decoderByName(n string) *encoding.Decoder {
	n = normalizeCharsetName(n)
	for _, cm := range charmaps {
		name := normalizeCharsetName(cm.String())
		if name == n {
			return cm.NewDecoder()
		}
	}

	return nil
}

func normalizeCharsetName(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
