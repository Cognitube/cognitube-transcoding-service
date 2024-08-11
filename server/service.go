package server

type ICognitubeTranscodingService interface {
	TranscodeVideo(videoId string, url string)
	ExtractAudio(videoId string, url string)
}
