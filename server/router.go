package server

import (
	"github.com/gorilla/mux"
)

// SetupRoutes configures the HTTP routes for the server.
func SetupRoutes(r *mux.Router, TranscodingService ICognitubeTranscodingService) {
	TranscodingHandler := NewTranscodingHandler(TranscodingService)
	r.HandleFunc("/", Home).Methods("GET")
	r.HandleFunc("/api/v1/transcode", TranscodingHandler.Transcode).Methods("POST")
	r.HandleFunc("/api/v1/extract-audio", TranscodingHandler.ExtractAudio).Methods("POST")
}
