package server

type ICognitubeTranscodingService interface {
	TranscodeVideo(videoId string, url string, retryCount int)
	ExtractAudio(videoId string, url string, retryCount int)
}
