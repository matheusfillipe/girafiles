package api

import "strings"

var SUP_MIMETYPES_PRE = []string{
	"text",
	"image",
	"audio",
	"video",
}

var SUP_MIMETYPES = []string{
	"application/pdf",
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
