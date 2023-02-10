package api

import (
	"github.com/russross/blackfriday"
)

func markdownToHtml(markdown string) string {
	unsafe := blackfriday.Markdown([]byte(markdown), renderer, extensions)
	return string(policy.SanitizeBytes(unsafe))
}
