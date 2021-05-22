package content

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

func TestAmpImg(t *testing.T) {
	assert := assert.New(t)

	html := `<figure>
		<amp-img src="image.jpg" srcset="one.jpg 704w, two.jpg 512w, three.jpg 310w">
			<img src="picture.jpg" srcset="large.jpg 704w, medium.jpg 512w, small.jpg 310w" />
		</amp-img>
		<figcaption>Caption</figcaption>
	</figure>`
	expect := `<figure>
		<img src="image.jpg" srcset="one.jpg 704w, two.jpg 512w, three.jpg 310w"/>
		<figcaption>Caption</figcaption>
	</figure>`
	d := doc(html)
	convertAmpImg(d)

	assert.Equal(expect, str(d))
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

func TestResolveSrcset(t *testing.T) {
	assert := assert.New(t)

	html := "<picture><img src=\"default.jpg\"/></picture>"
	d := doc(html)
	resolveSrcset(d)
	assert.Equal(html, str(d))

	d = doc(`<img src="foo.jpg" srcset="small.jpg 100w, large.jpg 200w"/>`)
	resolveSrcset(d)
	img := d.Selection.Find("img").First()
	src, _ := img.Attr("src")
	assert.Equal("large.jpg", src)
}

func TestSrcsetMedia(t *testing.T) {
	assert := assert.New(t)

	html := "<picture><img src=\"default.jpg\"/></picture>"
	d := doc(html)
	resolvePicture(d)
	assert.Equal(html, str(d))

	html = `<picture>
        <source media="(max-width: 400px)" data-srcset="small.png" />
        <source media="(max-width: 1000px)" data-srcset="large.png" />
        <img src="default.jpg" />
    </picture>`
	d = doc(html)
	resolvePicture(d)
	assert.Equal("<img src=\"large.png\"/>", strings.TrimSpace(str(d)))
}

func TestParseMediaQueryWidth(t *testing.T) {
	assert := assert.New(t)
	var mq mediaQuery

	// max, min
	mq = parseMediaQueryWidth("(width:  1024px)")
	assert.False(mq.IsEmpty())
	assert.Equal(1024, mq.width)
	assert.Equal(0, mq.minWidth)
	assert.Equal(0, mq.maxWidth)

	mq = parseMediaQueryWidth("(max-width:  1024px)")
	assert.False(mq.IsEmpty())
	assert.Equal(1024, mq.maxWidth)

	mq = parseMediaQueryWidth("(min-width:  1024px)")
	assert.False(mq.IsEmpty())
	assert.Equal(1024, mq.minWidth)

	// space is optional
	mq = parseMediaQueryWidth("(width:1024px)")
	assert.False(mq.IsEmpty())

	// no match w/o "...px"
	mq = parseMediaQueryWidth("(width:  1024)")
	assert.True(mq.IsEmpty())

	// multiple queries DO NOT WORK
	// it should capture the first query and ignore the second one
	mq = parseMediaQueryWidth("(max-width:  1024px) or (min-width: 500px)")
	assert.False(mq.IsEmpty())
	assert.Equal(1024, mq.maxWidth)

	// may include stuff we don'T understand
	mq = parseMediaQueryWidth("(width:  1024px) and (orientation: landscape)")
	assert.False(mq.IsEmpty())
}
