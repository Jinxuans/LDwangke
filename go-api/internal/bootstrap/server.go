package bootstrap

import (
	"net/http"

	"go-api/internal/config"
)

// NewHTTPServer 构造 HTTP Server。
func NewHTTPServer(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: handler,
	}
}
