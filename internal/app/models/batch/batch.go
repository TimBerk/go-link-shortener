// Package batch модуль для работы с моделями пачек данных
package batch

//go:generate easyjson -all -snake_case batch.go

// ItemRequest - параметры записи запроса с идентификатором и ссылкой
type ItemRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

//easyjson:json
type BatchRequest []ItemRequest

// ItemResponse - параметры записи ответа с идентификатором и короткой ссылкой
type ItemResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

//easyjson:json
type BatchResponse []ItemResponse
