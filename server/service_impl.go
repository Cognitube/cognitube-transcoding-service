package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"

	"cognitube.com/transcoding-service/azure"
	"cognitube.com/transcoding-service/env"
	"cognitube.com/transcoding-service/publish"
	"cognitube.com/transcoding-service/result"
)

type FFProbeOutput struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

type CongnitubeTranscodingService struct {
	resultPublisher publish.TranscodingPublisher
	blobClient      *azure.BlobClient
}

func NewCognitubeTranscodingService() ICognitubeTranscodingService {
	return &CongnitubeTranscodingService{
		resultPublisher: publish.NewKafkaTranscodingPublisher(),
		blobClient:      azure.NewBlobClient(),
	}
}

func (c *CongnitubeTranscodingService) hasAudio(file os.File) (bool, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "a:0", "-show_entries", "stream=codec_name", "-of", "default=noprint_wrappers=1:nokey=1", file.Name())
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return string(out) != "", nil
}

func (c *CongnitubeTranscodingService) processVidAsync(videoID string, url string, retryCount int) {
	retry := -1
	success := false
	var err error
	var resultURL string
	var duration float64
	var rslt *result.TranscodingResult

	log.Println("Transcoding video with ID:", videoID)

	for retry < env.GetInstance().TranscodingMaxRetry {
		duration, resultURL, err = c.processVideo(url)
		if err != nil {
			retry += 1
			log.Printf("Error while transcoding video: %v. Retrying...", err)
			continue
		}

		success = true
		break
	}

	if success {
		rslt = &result.TranscodingResult{
			VideoID:       videoID,
			VideoDuration: duration,
			VideoURL:      resultURL,
			Success:       true,
			Error:         "",
			RetryCount:    retryCount,
		}
	} else {
		rslt = &result.TranscodingResult{
			VideoID:       videoID,
			VideoURL:      "",
			VideoDuration: 0,
			Success:       false,
			Error:         err.Error(),
			RetryCount:    retryCount,
		}
	}

	err = c.resultPublisher.PublishTranscodingResult(rslt)
	if err != nil {
		log.Println("failed to publish result:", err)
	}

	log.Println("Transcoding video with ID:", videoID, "completed")
}

func (c *CongnitubeTranscodingService) extractAudio(url string) (string, error) {
	bytes, err := c.blobClient.DownloadFromBlob(url)
	if err != nil {
		log.Println("failed to download file:", err)
		return "", err
	}

	srcFile, err := os.CreateTemp("", "temp-")
	if err != nil {
		log.Println("Error while creating temp file:", err.Error())
		return "", err
	}
	defer srcFile.Close()
	defer os.Remove(srcFile.Name())

	_, err = srcFile.Write(bytes)
	if err != nil {
		log.Println("Error while writing to temp file:", err.Error())
		return "", err
	}

	videoHasAudio, err := c.hasAudio(*srcFile)
	if err != nil {
		log.Println("Error while checking if video has audio:", err.Error())
		return "", err
	}

	if !videoHasAudio {
		return "", fmt.Errorf("Video does not contain an audio track")
	}

	targetFile, err := os.CreateTemp("", "audio-*.oga")
	if err != nil {
		log.Println("Error while creating temp file:", err.Error())
		return "", err
	}
	defer targetFile.Close()
	defer os.Remove(targetFile.Name())

	command := []string{
		"ffmpeg",
		"-y",                 // -y: 无询问覆盖输出文件
		"-i", srcFile.Name(), // -i: 输入文件路径
		"-vn",             // -vn: 不包含视频流
		"-c:a", "libopus", // -c:a: 音频编解码器，使用 Opus
		"-b:a", env.GetInstance().AudioBitRate + "k", // -b:a: 音频比特率
		"-ac", "1", // -ac: 音频通道，设置音频通道数为 1 (单声道)
		"-ar", env.GetInstance().AudioSamplingRate, // -ar: 音频采样率
		"-threads", "0", // -threads: 自动确定使用的线程数
		targetFile.Name(), // 输出文件路径
	}

	cmd := exec.Command(command[0], command[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Println("ffmpeg command failed:", err.Error())
		return "", err
	}

	log.Println("ffmpeg command executed successfully")

	data, err := io.ReadAll(targetFile)
	if err != nil {
		log.Println("failed to read file:", err.Error())
		return "", err
	}

	blobURL, err := c.blobClient.UploadBlob(env.GetInstance().AudioContainerName, generateRandomFilename("audio-"), data)
	if err != nil {
		log.Println("failed to upload file:", err)
		return "", err
	}

	return blobURL, nil
}

func (c *CongnitubeTranscodingService) extractVidAudioAsync(videoID string, url string, retryCount int) {
	retry := -1
	success := false
	var err error
	var resultURL string
	var rslt *result.AudioExtractionResult

	log.Println("Extracting audio for video with ID:", videoID)

	for retry < env.GetInstance().TranscodingMaxRetry {
		resultURL, err = c.extractAudio(url)
		if err != nil {
			if err.Error() == "Video does not contain an audio track" {
				log.Println("Video does not contain an audio track. Skipping keyword extraction.")
				break
			}
			retry += 1
			log.Printf("Error while extracting audio for video: %v. Retrying...", err)
			continue
		}

		success = true
		break
	}

	if success {
		rslt = &result.AudioExtractionResult{
			VideoID:    videoID,
			AudioURL:   resultURL,
			Success:    true,
			Error:      "",
			RetryCount: retryCount,
		}
	} else {
		rslt = &result.AudioExtractionResult{
			VideoID:    videoID,
			AudioURL:   "",
			Success:    false,
			Error:      err.Error(),
			RetryCount: retryCount,
		}
	}

	err = c.resultPublisher.PublishAudioExtractionResult(rslt)
	if err != nil {
		log.Println("failed to publish result:", err)
		return
	}

	log.Println("Audio extraction for video with ID:", videoID, "completed")
}

func (c *CongnitubeTranscodingService) ExtractAudio(videoID string, url string, retryCount int) {
	go c.extractVidAudioAsync(videoID, url, retryCount)
}

func (c *CongnitubeTranscodingService) TranscodeVideo(videoID string, url string, retryCount int) {
	go c.processVidAsync(videoID, url, retryCount)
}

func (c *CongnitubeTranscodingService) processVideo(url string) (float64, string, error) {
	bytes, err := c.blobClient.DownloadFromBlob(url)
	if err != nil {
		return 0, "", err
	}

	srcFile, err := os.CreateTemp("", "temp-")
	if err != nil {
		log.Println("Error while creating temp file:", err.Error())
		return 0, "", err
	}
	defer srcFile.Close()
	defer os.Remove(srcFile.Name())

	_, err = srcFile.Write(bytes)
	if err != nil {
		log.Println("Error while writing to temp file:", err.Error())
		return 0, "", err
	}

	targetFile, err := os.CreateTemp("", "processed-*.mp4")
	if err != nil {
		log.Println("Error while creating temp file:", err.Error())
		return 0, "", err
	}
	defer targetFile.Close()
	defer os.Remove(targetFile.Name())

	command := []string{
		"ffmpeg",
		"-y",                 // -y: 无询问覆盖输出文件
		"-i", srcFile.Name(), // -i: 输入文件路径
		"-c:v", "libx264", // -c:v: 视频编解码器，使用 H.264
		"-b:v", "800k", // -b:v: 视频比特率，设置视频比特率为 800 kbps
		"-r", "30", // -r: 帧率，设置帧率为 30 fps
		"-c:a", "aac", // -c:a: 音频编解码器，使用 AAC
		"-b:a", "128k", // -b:a: 音频比特率，设置音频比特率为 128 kbps
		"-ac", "2", // -ac: 音频通道，设置音频通道数为 2 (立体声)
		"-ar", "44100", // -ar: 音频采样率，设置音频采样率为 44100 Hz
		"-threads", "0", // -threads: 自动确定使用的线程数
		targetFile.Name(), // 输出文件路径
	}

	cmd := exec.Command(command[0], command[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Println("ffmpeg command failed:", err.Error())
		return 0, "", err
	}

	log.Println("ffmpeg command executed successfully")

	data, err := io.ReadAll(targetFile)
	if err != nil {
		log.Println("failed to read file:", err.Error())
		return 0, "", err
	}

	blobURL, err := c.blobClient.UploadBlob(env.GetInstance().VideoContainerName, generateRandomFilename("processed-"), data)
	if err != nil {
		log.Println("failed to upload file:", err)
		return 0, "", err
	}

	shortURL, err := c.blobClient.ExtractContainerAndBlob(blobURL)
	if err != nil {
		log.Println("failed to extract container and blob:", err)
		return 0, "", err
	}

	duration, err := getVideoDuration(targetFile.Name())
	if err != nil {
		log.Println("failed to get video duration:", err)
		return 0, "", err
	}

	return duration, shortURL, nil
}

func getVideoDuration(filename string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "json", filename)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var ffprobeOutput FFProbeOutput
	if err := json.Unmarshal(out, &ffprobeOutput); err != nil {
		return 0, err
	}

	duration, err := strconv.ParseFloat(ffprobeOutput.Format.Duration, 64)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

func generateRandomFilename(prefix string) string {
	randBytes := make([]byte, 16) // 生成 16 字节的随机数
	if _, err := rand.Read(randBytes); err != nil {
		log.Fatal("Failed to generate random bytes:", err)
	}
	return fmt.Sprintf("%s%s.mp4", prefix, hex.EncodeToString(randBytes)) // 返回前缀和随机字符串组合的文件名
}
