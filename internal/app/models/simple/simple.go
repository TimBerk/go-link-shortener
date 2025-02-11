// simple.go
package simple

//go:generate easyjson -all -snake_case simple.go

type RequestJSON struct {
	URL string `json:"url"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}
