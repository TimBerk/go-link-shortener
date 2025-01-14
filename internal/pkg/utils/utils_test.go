package utils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckParamInHeaderParam(t *testing.T) {
	tests := []struct {
		name        string
		headers     map[string]string
		headerParam string
		needParam   string
		want        bool
	}{
		{
			name:        "Empty header",
			headers:     map[string]string{},
			headerParam: "Content-type",
			needParam:   "text/plain",
			want:        false,
		},
		{
			name:        "Header without necessary header param",
			headers:     map[string]string{"x-frame-options": "SAMEORIGIN"},
			headerParam: "Content-type",
			needParam:   "text/plain",
			want:        false,
		},
		{
			name:        "Header with necessary param",
			headers:     map[string]string{"Content-type": "text/plain;charset=utf-8"},
			headerParam: "Content-type",
			needParam:   "text/plain",
			want:        true,
		},
		{
			name:        "Header without necessary param",
			headers:     map[string]string{"Content-type": "charset=utf-8"},
			headerParam: "Content-type",
			needParam:   "text/plain",
			want:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			assert.Equal(
				t,
				CheckParamInHeaderParam(req, test.headerParam, test.needParam),
				test.want,
				"CheckParamInHeaderParam(%s, %s) mast return %t",
				test.headerParam,
				test.needParam,
				test.want,
			)

		})
	}

}
