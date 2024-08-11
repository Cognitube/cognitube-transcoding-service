package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type TranscodingHandler struct {
	Service ICognitubeTranscodingService
}

func NewTranscodingHandler(service ICognitubeTranscodingService) *TranscodingHandler {
	return &TranscodingHandler{Service: service}
}

func Home(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
	/ - Home
	/api/v1/transcode - [POST] Transcode a video
	/api/v1/extract-audio - [POST] Extract audio from a video
	`))
}

func (t *TranscodingHandler) ExtractAudio(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		VideoID  string `json:"videoId"`
		VideoURL string `json:"videoUrl"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	err = json.Unmarshal(body, &reqData)
	log.Printf("Received request: %+v\n", reqData)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	t.Service.ExtractAudio(reqData.VideoID, reqData.VideoURL)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Extracting audio"))
}

func (t *TranscodingHandler) Transcode(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		VideoID  string `json:"videoId"`
		VideoURL string `json:"videoUrl"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	err = json.Unmarshal(body, &reqData)
	log.Printf("Received request: %+v\n", reqData)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	t.Service.TranscodeVideo(reqData.VideoID, reqData.VideoURL)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Transcoding"))
}
