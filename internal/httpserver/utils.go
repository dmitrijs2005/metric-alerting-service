package httpserver

import "strings"

func ContentTypeIsCompressable(contentType string) bool {
	return contentType == "application/json" || strings.HasPrefix(contentType, "text/html")
}
