package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type JSONErrorResponse struct {
	Error string `json:"error"`
}

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

func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	errorResponse := JSONErrorResponse{
		Error: message,
	}

	log.Printf("JSONError - (%s): [(%s)] %d", w.Header(), message, statusCode)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(errorResponse)
}
