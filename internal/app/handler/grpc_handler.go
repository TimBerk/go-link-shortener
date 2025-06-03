package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/TimBerk/go-link-shortener/api/gen"
	"github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
)

// GRPCHandler - структура для хранения настроек и обработчиков данных
type GRPCHandler struct {
	gen.UnimplementedShortenerServer
	handler *Handler
}

// NewGRPCHandler - инициализация нового gRPC-обработчика на основании переаданного обработчика
func NewGRPCHandler(h *Handler) *GRPCHandler {
	return &GRPCHandler{handler: h}
}

// ShortenURL обрабатывает запрос на сокращение URL
func (h *GRPCHandler) ShortenURL(ctx context.Context, req *gen.ShortenURLRequest) (*gen.ShortenURLResponse, error) {
	shortURL, err := h.handler.store.AddURL(ctx, req.Url, req.UserId)
	exists := errors.Is(err, store.ErrLinkExist)
	if err != nil && !exists {
		return nil, status.Errorf(codes.Internal, "failed to add URL: %v", err)
	}

	fullShortURL := fmt.Sprintf("http://%s/%s", h.handler.cfg.ServerAddress, shortURL)
	addShortURL(req.UserId, fullShortURL, req.Url)

	return &gen.ShortenURLResponse{
		ShortUrl: fullShortURL,
		Exists:   exists,
	}, nil
}

// ShortenJSONURL обрабатывает запрос на сокращение URL в JSON формате
func (h *GRPCHandler) ShortenJSONURL(ctx context.Context, req *gen.ShortenJSONURLRequest) (*gen.ShortenJSONURLResponse, error) {
	shortURL, err := h.handler.store.AddURL(ctx, req.Url, req.UserId)
	exists := errors.Is(err, store.ErrLinkExist)
	if err != nil && !exists {
		return nil, status.Errorf(codes.Internal, "failed to add URL: %v", err)
	}

	fullShortURL := fmt.Sprintf("http://%s/%s", h.handler.cfg.ServerAddress, shortURL)
	addShortURL(req.UserId, fullShortURL, req.Url)

	return &gen.ShortenJSONURLResponse{
		Result: fullShortURL,
		Exists: exists,
	}, nil
}

// Redirect выполняет перенаправление по короткому URL
func (h *GRPCHandler) Redirect(ctx context.Context, req *gen.RedirectRequest) (*gen.RedirectResponse, error) {
	originalURL, exists, isDeleted := h.handler.store.GetOriginalURL(ctx, req.ShortUrl, req.UserId)
	if !exists {
		return nil, status.Errorf(codes.NotFound, "short URL not found")
	}
	if isDeleted {
		return nil, status.Errorf(codes.Unavailable, "URL is deleted")
	}

	return &gen.RedirectResponse{
		OriginalUrl: originalURL,
		Exists:      exists,
		IsDeleted:   isDeleted,
	}, nil
}

// Ping проверяет соединение с базой данных
func (h *GRPCHandler) Ping(ctx context.Context, req *gen.PingRequest) (*gen.PingResponse, error) {
	err := h.handler.store.Ping(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database ping failed: %v", err)
	}
	return &gen.PingResponse{Success: true}, nil
}

// ShortenBatch обрабатывает пакетное создание коротких URL
func (h *GRPCHandler) ShortenBatch(ctx context.Context, req *gen.ShortenBatchRequest) (*gen.ShortenBatchResponse, error) {
	batchReq := make(batch.BatchRequest, len(req.Urls))
	for i, url := range req.Urls {
		batchReq[i] = batch.ItemRequest{
			CorrelationID: url.CorrelationId,
			OriginalURL:   url.OriginalUrl,
		}
	}

	batchResp, err := h.handler.store.AddURLs(ctx, batchReq, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add batch URLs: %v", err)
	}

	resp := &gen.ShortenBatchResponse{
		Urls: make([]*gen.BatchResponse, len(batchResp)),
	}
	for i, url := range batchResp {
		fullShortURL := fmt.Sprintf("http://%s/%s", h.handler.cfg.ServerAddress, url.ShortURL)
		resp.Urls[i] = &gen.BatchResponse{
			CorrelationId: url.CorrelationID,
			ShortUrl:      fullShortURL,
		}
		addShortURL(req.UserId, fullShortURL, fullShortURL)
	}

	return resp, nil
}

// GetUserURLs обрабатывает ссылки пользователя
func (h *GRPCHandler) GetUserURLs(ctx context.Context, req *gen.GetUserURLsRequest) (*gen.GetUserURLsResponse, error) {
	urls, exists := userURLs[req.UserId]
	if !exists || len(urls) == 0 {
		return nil, status.Errorf(codes.NotFound, "no URLs found for user")
	}

	resp := &gen.GetUserURLsResponse{
		Urls: make([]*gen.UserURL, len(urls)),
	}
	for i, url := range urls {
		resp.Urls[i] = &gen.UserURL{
			ShortUrl:    url["short_url"],
			OriginalUrl: url["original_url"],
		}
	}

	return resp, nil
}

// DeleteURLs помечает URL как удаленные
func (h *GRPCHandler) DeleteURLs(ctx context.Context, req *gen.DeleteURLsRequest) (*gen.DeleteURLsResponse, error) {
	for _, shortURL := range req.ShortUrls {
		h.handler.urlChan <- store.URLPair{ShortURL: shortURL, UserID: req.UserId}
		logrus.WithFields(logrus.Fields{
			"shortURL": shortURL,
			"UserID":   req.UserId,
		}).Info("Deleted user link")
	}

	return &gen.DeleteURLsResponse{Success: true}, nil
}
