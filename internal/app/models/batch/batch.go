// batch.go
package batch

//go:generate easyjson -all -snake_case batch.go

type ItemRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

//easyjson:json
type BatchRequest []ItemRequest

type ItemResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

//easyjson:json
type BatchResponse []ItemResponse
