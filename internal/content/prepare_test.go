package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnwrapNoscript(t *testing.T) {
	assert := assert.New(t)

	d := doc("<p>foo <noscript><em>bar</em></noscript> baz</p>")
	unwrapNoscript(d)
	assert.Equal("<p>foo <em>bar</em> baz</p>", str(d))
}

func TestUnwrapDivs(t *testing.T) {
	assert := assert.New(t)

	html := `
	<div>
      <div>
        <div>
          <picture class="ts-picture js-picture ts-picture--copytext-l">
            <source type="image/png" media="(max-width: 420px)" src="https://www.tagesschau.de/res/assets/image/lazy-image-placeholder.jpg" data-srcset="https://www.tagesschau.de/multimedia/bilder/alex-barck-101~_v-mittelgross1x1.jpg"/>
            <source type="image/png" media="(max-width: 767px)" src="https://www.tagesschau.de/res/assets/image/lazy-image-placeholder.jpg" data-srcset="https://www.tagesschau.de/multimedia/bilder/alex-barck-101~_v-videoweb1x1l.jpg"/>
            <img class="ts-image js-image" title="DJ Alex Barck legt am Dienstag (18.05.2010) beim Verve Club auf. | picture-alliance / Xamax" alt="DJ Alex Barck legt am Dienstag (18.05.2010) beim Verve Club auf. | picture-alliance / Xamax" src="https://www.tagesschau.de/multimedia/bilder/alex-barck-101~_v-videowebl.jpg"/>
          </picture>
        </div>
      </div>
      <div>
        <p>Sonst am Plattenteller, jetzt im Team des Impfzentrums Tegel: DJ Alex Barck.</p>
      </div>
    </div>`

	d := doc(html)
	unwrapDivs(d)
	assert.Equal(2, d.Find("div").Length())

}
