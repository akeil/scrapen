package content

// contains HTML elements which we do not want in the output.
// But we DO want to keep their text content.
var unwrap = []string{
	"span", "div",
	"article", "section", "summary",
	"address",
	"main", "footer", "header",
	"hgroup",
	"data",
	"dfn",
	// deprecated elements
	"acronym", "basefont", "big", "blink", "center",
	"content", "font", "listing",
	"marquee", "nobr", "plaintext", "spacer",
	"strike", "tt",
	"picture",
}

// Tells if the given tag is "gray", e.g. we do want to keep its text content
// but we do NOT want this tag in our output.
func isGraylisted(tag string) bool {
	for _, greylisted := range unwrap {
		if tag == greylisted {
			return true
		}
	}

	return false
}

// keep `id` for all elements?
// to support internal links?
var attrWhitelist = map[string][]string{
	"img": []string{"src", "width", "height", "alt", "title"},
	"a":   []string{"href", "title"}, // rel?
	//"svg":   []string{"xmlns", "viewBox", "version", "x", "y", "style"},
	//"path":   []string{"d"},
}

// Tell if a given attribute name is allowed for a given element.
func isWhitelistedAttr(tag, attr string) bool {
	whitelist, ok := attrWhitelist[tag]
	if !ok {
		return false
	}

	for _, whitelisted := range whitelist {
		if attr == whitelisted {
			return true
		}
	}

	return false
}

// whitelist contains HTML elements which we want to keep.
var whitelist = []string{
	"#document", "html", "body",
	"p",
	"a",
	"h1", "h2", "h3", "h4", "h5", "h6",
	"br", "hr",
	"b", "u", "i", "s",
	"em", "strong", "small",
	"sub", "sup",
	"abbr",
	"del", "ins",
	"aside",
	"ul", "ol", "li",
	"dl", "dd", "dt",
	"table", "thead", "tbody", "tfoot", "caption", "tr", "th", "td", "colgroup", "col",
	"code", "pre", "kbd", "sample", "var",
	"mark", "q",
	"rp", "rt", "rtc", "ruby",
	"blockquote", "cite",
	"img",
	"figure", "figcaption",
	"bdi", "bdo",
	"time",
	"wbr",
	// "audio", "video", "track", "source",
	// embed, iframe,
	// object, param,
	// picture, source
	// svg, path, g
	// nav  <-- drop as it likely contains irrelevant content
}

// Tell if the given element is one that we wish to see in our clen content.
func isWhitelistedTag(tag string) bool {
	for _, whitelisted := range whitelist {
		if tag == whitelisted {
			return true
		}
	}
	// tags we want empty, but NOT removed
	return isGraylisted(tag)
}

// see
// https://developer.mozilla.org/en-US/docs/Web/HTML/Block-level_elements
var blocklevel = []string{
	"p",
	"h1", "h2", "h3", "h4", "h5", "h6",
	"hr",
	"aside",
	"ul", "ol", "li",
	"dl", "dd", "dt",
	"blockquote",
	"pre",
	"figure", "figcaption",
	// caption, th, td are actually not block-levels - but still can trim space
	"table", "caption", "th", "td",
}

// tell if the given element is a block-level element.
// works for whitelisted tags only.
func isBlocklevel(tag string) bool {
	for _, bl := range blocklevel {
		if tag == bl {
			return true
		}
	}

	return false
}
