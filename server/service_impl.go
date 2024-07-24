package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

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

func (c *CongnitubeTranscodingService) processVidAsync(videoID string, url string) {
	retry := -1
	success := false
	var err error
	var resultURL string
	var rslt *result.Result

	log.Println("Transcoding video with ID:", videoID)

	for retry < env.GetInstance().TranscodingMaxRetry {
		resultURL, err = c.processVideo(url)
		if err != nil {
			retry += 1
			log.Printf("Error while transcoding video: %v. Retrying...", err)
			continue
		}

		success = true
		break
	}

	if success {
		rslt = &result.Result{
			VideoID:  videoID,
			VideoURL: resultURL,
			Success:  true,
			Error:    "",
		}
	} else {
		rslt = &result.Result{
			VideoID:  videoID,
			VideoURL: "",
			Success:  false,
			Error:    err.Error(),
		}
	}

	err = c.resultPublisher.PublishTranscodingResult(rslt)
	if err != nil {
		log.Println("failed to publish result:", err)
	}
}

func (c *CongnitubeTranscodingService) TranscodeVideo(videoID string, url string) {
	go c.processVidAsync(videoID, url)
}

func (c *CongnitubeTranscodingService) processVideo(url string) (string, error) {
	bytes, err := c.blobClient.DownloadFromBlob(url)
	if err != nil {
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

	targetFile, err := os.CreateTemp("", "processed-*.mp4")
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
		return "", err
	}

	log.Println("ffmpeg command executed successfully")

	data, err := io.ReadAll(targetFile)
	if err != nil {
		log.Println("failed to read file:", err.Error())
		return "", err
	}

	blobURL, err := c.blobClient.UploadBlob(env.GetInstance().VideoContainerName, generateRandomFilename("processed-"), data)
	if err != nil {
		log.Println("failed to upload file:", err)
		return "", err
	}

	shortURL, err := c.blobClient.ExtractContainerAndBlob(blobURL)
	if err != nil {
		log.Println("failed to extract container and blob:", err)
		return "", err
	}

	return shortURL, nil
}

func generateRandomFilename(prefix string) string {
	randBytes := make([]byte, 16) // 生成 16 字节的随机数
	if _, err := rand.Read(randBytes); err != nil {
		log.Fatal("Failed to generate random bytes:", err)
	}
	return fmt.Sprintf("%s%s.mp4", prefix, hex.EncodeToString(randBytes)) // 返回前缀和随机字符串组合的文件名
}
