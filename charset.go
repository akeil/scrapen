package scrapen

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

// readUTF8 reads the response body into a UTF-8 string
func readUTF8(r io.Reader, h http.Header) (string, error) {
	cs := charsetFromHeader(h)
	fmt.Printf("Found charset %q from header\n", cs)
	// TODO: we can also obtain the charset from <meta> tag
	// ... and XML declaration

	if normalizeCharsetName(cs) != "utf-8" && cs != "" {
		dec := decoderByName(cs)
		if dec != nil {
			fmt.Printf("Found decoder for charset %q\n", cs)
			r = dec.Reader(r)
		} else {
			fmt.Printf("Could not find decoder for charset %q, assume UTF-8\n", cs)
		}
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func charsetFromHeader(h http.Header) string {
	contentType, ok := h["Content-Type"]
	if !ok {
		return ""
	}

	for _, s := range contentType {
		_, params, err := mime.ParseMediaType(s)
		if err != nil {
			fmt.Printf("error parsing media type: %v", err)
			return ""
		}
		cs := params["charset"]
		if cs != "" {
			return cs
		}
	}

	return ""
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
