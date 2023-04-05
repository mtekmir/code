package server

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Server holds the dependencies for the http server
type Server struct {
	Db     *sql.DB
	Router chi.Router
}

// New Initiates a new server.
func New() *Server {
	s := &Server{
		Router: chi.NewRouter(),
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
	log.Println("Server starting at port 8080")
	if err := http.ListenAndServe(":8080", s); err != nil {
		return err
	}

	return nil
}
