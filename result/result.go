package result

type Result struct {
	Success       bool    `json:"success"`
	Error         string  `json:"error"`
	VideoID       string  `json:"videoId"`
	VideoURL      string  `json:"videoUrl"`
	VideoDuration float64 `json:"videoDuration"`
}
