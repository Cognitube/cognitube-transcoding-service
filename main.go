package main

import (
	"os"

	"cognitube.com/transcoding-service/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9032"
	}

	transcodingServer := server.NewCognitubeTranscodingServer()
	transcodingServer.StartListening(port)
}
