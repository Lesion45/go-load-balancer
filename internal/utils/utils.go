package utils

import (
	"net/http"

	"github.com/Lesion45/go-load-balancer/internal/backend"
)

func CreateTestBackend(url string, healthy bool, connections int64) *backend.SimpleBackend {
	b := &backend.SimpleBackend{URL: url}
	b.SetStatus(healthy)
	for i := int64(0); i < connections; i++ {
		b.AddActiveConnection()
	}
	return b
}

type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *StatusRecorder) WriteHeader(code int) {
	r.Status = code
	r.ResponseWriter.WriteHeader(code)
}
