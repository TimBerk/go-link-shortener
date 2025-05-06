// Package compress выступает в роли обработчика запросов
// для осуществления сжатия и декодирования контента для gzip
package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/TimBerk/go-link-shortener/internal/pkg/utils"

	"github.com/sirupsen/logrus"
)

// gzipWriter - параметры для записи контента в gzip-формате
type gzipWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// gzipReader - параметры для чтения контента в gzip-формате
type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// newCompressWriter - создание нового обработчика для записи контента в gzip-формате
func newCompressWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header - получение заголовка из обработчика записи
func (c *gzipWriter) Header() http.Header {
	return c.w.Header()
}

// Write - запись данных с помощью обработчика записи
func (c *gzipWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader - запись данных в заголовок с помощью обработчика записи
func (c *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close - закрытие обработчика записи
func (c *gzipWriter) Close() error {
	return c.zw.Close()
}

// newCompressReader - создание нового обработчика для чтения контента в gzip-формате
func newCompressReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read - чтения данных из обработчика чтения
func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close - закрытие обработчика чтения
func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// GzipMiddleware - обработчик запросов для осуществления сжатия/декодирования контента
// При наличии заголовка Accept-Encoding со значением gzip осуществляется сжатие данных
// При наличии заголовка Content-Encoding со значением gzip осуществляется декодирование данных
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := newCompressWriter(w)
			ow = cw
			defer utils.CloseWithLog(cw, "Error closing CompressWriter")
		}

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"header": w.Header(),
					"err":    err,
				}).Error("Header error for Content-Encoding")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer utils.CloseWithLog(cr, "Error closing CompressReader")
		}

		next.ServeHTTP(ow, r)
	})
}
