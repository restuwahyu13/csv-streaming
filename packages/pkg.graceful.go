package packages

import (
	"crypto/rand"
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ory/graceful"
)

type GracefulConfig struct {
	Handler *chi.Mux
	Port    string
}

func Graceful(Handler func() *GracefulConfig) error {
	h := Handler()
	secure := false

	if os.Getenv("GO_ENV") == "development" {
		secure = true
	}

	server := http.Server{
		Handler:        h.Handler,
		Addr:           ":" + h.Port,
		WriteTimeout:   time.Duration(time.Minute * time.Duration(15)),
		ReadTimeout:    time.Duration(time.Minute * time.Duration(3)),
		IdleTimeout:    time.Duration(time.Minute * time.Duration(1)),
		MaxHeaderBytes: 1e+9,
		TLSConfig: &tls.Config{
			Rand:               rand.Reader,
			InsecureSkipVerify: secure,
		},
	}

	return graceful.Graceful(server.ListenAndServe, server.Shutdown)
}
