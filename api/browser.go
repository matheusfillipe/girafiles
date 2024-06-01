package api

import "strings"

var SUP_MIMETYPES_PRE = []string{
	"text",
	"image",
	"audio",
	"video",
}

var SUP_MIMETYPES = []string{
	"application/json",
	"application/rss+xml",
	"application/xhtml+xml",
	"application/xml",
	"application/ogg",
	"application/pdf",
	"audio/mpeg",
	"audio/ogg",
	"audio/wav",

	"font/woff",
	"font/woff2",

	"image/bmp",
	"image/gif",
	"image/jpeg",
	"image/png",
	"image/svg+xml",
	"image/webp",
	"image/x-icon",

	"text/css",
	"text/html",
	"text/javascript",
	"text/plain",
	"text/xml",

	"video/mp4",
	"video/ogg",
	"video/webm",
}

func isSupportedMimetype(m string) bool {
	for _, s := range SUP_MIMETYPES {
		if s == m {
			return true
		}
	}
	for _, s := range SUP_MIMETYPES_PRE {
		if strings.Split(m, "/")[0] == s {
			return true
		}
	}
	return false
}
