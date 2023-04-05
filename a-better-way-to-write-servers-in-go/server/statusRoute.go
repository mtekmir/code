package server

import (
	"net/http"
)

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) error {
	// if err := s.Db.Ping(); err != nil {
	// 	return err
	// }

	w.WriteHeader(200)
	return nil
}
