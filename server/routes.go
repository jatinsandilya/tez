package server

import (
	_ "bitbucket.org/apps_scootsy/raas/docs" // docs is generated by Swag CLI, you have to import it.
	"github.com/gorilla/mux"
)

func (s *Server) setupRoutes(r *mux.Router) {

	// TODO : Add these to constants
	r.HandleFunc("/cache", Logger(s.handleSetKey)).Methods("POST")
	r.HandleFunc("/cache/{key}", Logger(s.handleGetKey)).Methods("GET")
	r.HandleFunc("/cache/{key}", Logger(s.handleDeleteKey)).Methods("DELETE")
	r.HandleFunc("/cache/pattern/{pattern}", Logger(s.handleDeleteKeysWithPattern)).Methods("DELETE")

	// TODO : Yet to develop below features
	r.HandleFunc("/cache/{key}", Logger(s.handleUpdateKey)).Methods("PUT")
	r.HandleFunc("/cache/{key}/{jsonpath}", Logger(s.handleUpdateSubKey)).Methods("PATCH")
	r.HandleFunc("/cache/pattern/{keyPattern}", s.handleGetValuesWithPattern).Methods("GET")
	r.HandleFunc("/cache/bulk", s.handleSetKeysInBulk).Methods("POST")
	r.HandleFunc("/cache/valuePattern/{ValuePattern}", s.handleGetKeysWithPattern).Methods("GET")
}
