package server

import "github.com/go-chi/chi/v5"

// SetupRoutes sets up the routes and middlewares.
func (s *Server) SetupRoutes() {
	r := chi.NewRouter()

	// ...
	r.Method("GET", "/status", handler(s.handleStatus))

	s.httpSrv.Handler = r
}
