package scrapen

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/google/brotli/go/cbrotli"
)

// br = Brotli
// compress = LZW
const supportedCompressions = "gzip, deflate"

func decompressed(r io.Reader, h http.Header) (io.Reader, error) {
	// identity
	f := func(r io.Reader) (io.Reader, error) {
		return r, nil
	}

	enc := h.Get("Content-Encoding")
	enc = strings.ToLower(enc)

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
