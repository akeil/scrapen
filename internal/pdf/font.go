package pdf

const (
	regular = 1 << iota
	bold
	italic
	underline
)

func fontStyle(code int) string {
	s := ""
	if code&bold != 0 {
		s += "B"
	}

	if code&italic != 0 {
		s += "I"
	}

	if code&underline != 0 {
		s += "U"
	}

	return s
}
