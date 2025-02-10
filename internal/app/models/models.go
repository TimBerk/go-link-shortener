// models.go
package models

//go:generate easyjson -all -snake_case models.go

type RequestJSON struct {
	URL string `json:"url"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}
