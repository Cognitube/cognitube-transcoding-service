package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type WebServer interface {
	StartListening(port string)
}

type CongitubeTranscodingServer struct {
	TranscodingService ICognitubeTranscodingService
}

func (c *CongitubeTranscodingServer) StartListening(port string) {
	router := mux.NewRouter()

	SetupRoutes(router, c.TranscodingService)

	log.Println("Server started at port ", port)

	http.ListenAndServe(":"+port, router)
}

func NewCongitubeTranscodingServer() WebServer {
	return &CongitubeTranscodingServer{
		TranscodingService: NewCognitubeTranscodingService(),
	}
}
