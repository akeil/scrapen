package assets

import "strings"

// minimal implementation to return file extension for supported content types
func fileExt(mime string) string {
	switch strings.ToLower(mime) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/svg+xml", "image/svg":
		return ".svg"
	default:
		return ""
	}
}
