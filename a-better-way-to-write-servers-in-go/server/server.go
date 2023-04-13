package server

import (
	"database/sql"
	"fmt"
	"github/mtekmir/a-server/config"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Server holds the dependencies for the http server
type Server struct {
	Db      *sql.DB
	Router  chi.Router
	httpSrv *http.Server
}

// New Initiates a new server.
func New(conf config.Config) *Server {
	s := &Server{
		httpSrv: &http.Server{
			Addr:           fmt.Sprintf(":%s", conf.Port),
			ReadTimeout:    conf.ReadTimeout,
			WriteTimeout:   conf.WriteTimeout,
			IdleTimeout:    conf.IdleTimeout,
			MaxHeaderBytes: conf.MaxHeaderBytes,
		},
	}
	s.SetupRoutes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

// Start starts the server. On local env it servers over http.
// On prod and test envs it configures autocert and serves over https.
func (s *Server) Start() error {
	log.Println("Server starting")
	if err := s.httpSrv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
