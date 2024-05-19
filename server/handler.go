package server

import "net/http"

type HTTPHandler struct {
	mux *http.ServeMux
}

// ServeHTTP implements http.Handler.
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

var _ http.Handler = (*HTTPHandler)(nil)
