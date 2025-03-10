package httpserver

import "strings"

func ContentTypeIsCompressable(contentType string) bool {
	return contentType == "application/json" || strings.HasPrefix(contentType, "text/html")
}

func int64Ptr(i int64) *int64 {
	return &i
}
func float64Ptr(f float64) *float64 {
	return &f
}
