package utils

import (
	"log"
	"net/http"
	"strings"
)

func CheckParamInHeaderParam(r *http.Request, headerParam string, needParam string) bool {
	headerValue := r.Header.Get(headerParam)
	if headerValue == "" {
		log.Printf("Header param (%s) is empty.", headerParam)
		return false
	}

	parts := strings.Split(headerValue, ";")
	if len(parts) < 1 {
		log.Printf("Header param (%s) is malformed.", headerParam)
		return false
	}

	mediaType := strings.TrimSpace(parts[0])
	if mediaType == needParam {
		return true
	}

	log.Printf("Header param (%s) has not param (%s).", headerParam, needParam)
	return false
}
