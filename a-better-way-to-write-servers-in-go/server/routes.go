package server

// SetupRoutes sets up the routes and middlewares.
func (s *Server) SetupRoutes() {
	r := s.Router

	// ...
	r.Method("GET", "/status", handler(s.handleStatus))
}
