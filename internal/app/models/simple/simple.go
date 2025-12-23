// Package simple содержит модели запросов/ответов для JSON формата
package simple

//go:generate easyjson -all -snake_case simple.go

// RequestJSON описывает параметры запроса
type RequestJSON struct {
	URL string `json:"url"`
}

// ResponseJSON описывает параметры ответа
type ResponseJSON struct {
	Result string `json:"result"`
}

// StatsRequestJSON описывает параметры ответа для статистики
type StatsRequestJSON struct {
	URLs  int64 `json:"urls"`
	Users int64 `json:"users"`
}
