package fetch

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/brotli/go/cbrotli"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

// br = Brotli
// compress = LZW
const supportedCompressions = "gzip, deflate, br"

func decompressed(t *pipeline.Task, r io.Reader, h http.Header) (io.Reader, error) {
	// identity
	f := func(r io.Reader) (io.Reader, error) {
		return r, nil
	}

	enc := h.Get("Content-Encoding")
	enc = strings.ToLower(enc)

	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "fetch",
	}).Info(fmt.Sprintf("Require decompression for %q", enc))

	switch enc {
	case "gzip", "x-gzip":
		f = func(r io.Reader) (io.Reader, error) {
			return gzip.NewReader(r)
		}
	case "deflate":
		f = func(r io.Reader) (io.Reader, error) {
			return flate.NewReader(r), nil
		}
	case "br":
		f = func(r io.Reader) (io.Reader, error) {
			return cbrotli.NewReader(r), nil
		}
	}

	return f(r)
}
