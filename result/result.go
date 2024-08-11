package result

type TranscodingResult struct {
	Success       bool    `json:"success"`
	Error         string  `json:"error"`
	VideoID       string  `json:"videoId"`
	VideoURL      string  `json:"videoUrl"`
	VideoDuration float64 `json:"videoDuration"`
	RetryCount    int     `json:"retryCount"`
}

type AudioExtractionResult struct {
	Success    bool   `json:"success"`
	Error      string `json:"error"`
	VideoID    string `json:"videoId"`
	AudioURL   string `json:"audioUrl"`
	RetryCount int    `json:"retryCount"`
}
