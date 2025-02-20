package utils

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type JSONErrorResponse struct {
	Error string `json:"error"`
}

func CheckParamInHeaderParam(r *http.Request, headerParam string, needParam string) bool {
	headerValue := r.Header.Get(headerParam)
	if headerValue == "" {
		logrus.WithField("headerParam", headerParam).Error("Header param is empty")
		return false
	}

	parts := strings.Split(headerValue, ";")
	if len(parts) < 1 {
		logrus.WithField("headerParam", headerParam).Error("Header param is malformed")
		return false
	}

	mediaType := strings.TrimSpace(parts[0])
	if mediaType == needParam {
		return true
	}

	logrus.WithFields(logrus.Fields{
		"needParam":   needParam,
		"headerParam": headerParam,
	}).Error("Header param has not param")
	return false
}

func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	errorResponse := JSONErrorResponse{
		Error: message,
	}

	logrus.WithFields(logrus.Fields{
		"header":     w.Header(),
		"message":    message,
		"statusCode": statusCode,
	}).Error("JSON Error")

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(errorResponse)
}
