package clean

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPicture(t *testing.T) {
	assert := assert.New(t)

	html := "<picture><img src=\"default.jpg\"/></picture>"
	d := doc(html)
	resolvePicture(d)
	assert.Equal(html, str(d))

	html = `<picture>
        <source srcset="large.png 200w, small.png 100w" />
        <source srcset="dense.png 2x, standard.png 1x" />
        <source srcset="" />
        <img src="default.jpg" />
    </picture>`
	d = doc(html)
	resolvePicture(d)
	assert.Equal("<img src=\"large.png\"/>", strings.TrimSpace(str(d)))
}

func TestDataSrcset(t *testing.T) {
	assert := assert.New(t)

	html := "<picture><img src=\"default.jpg\"/></picture>"
	d := doc(html)
	resolvePicture(d)
	assert.Equal(html, str(d))

	html = `<picture>
        <source data-srcset="large.png 200w, small.png 100w" />
        <source data-srcset="dense.png 2x, standard.png 1x" />
        <source srcset="" />
        <img src="default.jpg" />
    </picture>`
	d = doc(html)
	resolvePicture(d)
	assert.Equal("<img src=\"large.png\"/>", strings.TrimSpace(str(d)))
}
