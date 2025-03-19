package httpserver

import (
	"bytes"
	"net/http"
)

type responseRecorder struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)                  // Записываем ответ в буфер
	return r.ResponseWriter.Write(b) // Передаем в оригинальный Writer
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func NewResponseRecorder(w http.ResponseWriter, b *bytes.Buffer) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, body: b}
}
