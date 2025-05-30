package internal

import (
	"bytes"
	"net/http"
)

type ResponseInterceptor struct {
	Body       *bytes.Buffer
	StatusCode int
	Written    bool
	headers    http.Header
}

func NewResponseInterceptor() *ResponseInterceptor {
	return &ResponseInterceptor{
		Body:       &bytes.Buffer{},
		StatusCode: 200,
		headers:    make(http.Header),
	}
}

func (r *ResponseInterceptor) Header() http.Header {
	return r.headers
}

func (r *ResponseInterceptor) Write(data []byte) (int, error) {
	if !r.Written {
		r.Written = true
	}
	return r.Body.Write(data)
}

func (r *ResponseInterceptor) WriteHeader(statusCode int) {
	if !r.Written {
		r.StatusCode = statusCode
		r.Written = true
	}
}

func (r *ResponseInterceptor) WriteTo(w http.ResponseWriter) {
	for key, values := range r.headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.WriteHeader(r.StatusCode)
	r.Body.WriteTo(w)
}
