package env

import (
	"os"
	"sync"
)

type Variables struct {
	Debug                     bool
	KafkaHost                 string
	KafkaPort                 string
	KafkaTranscodingTopic     string
	KafkaAudioExtractionTopic string
	BlobConnectString         string
	VideoContainerName        string
	TranscodingMaxRetry       int
	EventHubNamespace         string
	EventHubName              string
	EventHubConnectionString  string
	Username                  string
	AudioContainerName        string
	AudioExtractionMaxRetry   int
	AudioSamplingRate         string
	AudioBitRate              string
	KafkaBootstrapServers     string
}

var instance *Variables
var once sync.Once

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func loadValues() {
	instance = &Variables{
		Debug:                     os.Getenv("APPLICATION_DEBUG") == "true",
		KafkaHost:                 getEnvWithDefault("APPLICATION_KAFKA_HOST", "localhost"),
		KafkaPort:                 getEnvWithDefault("APPLICATION_KAFKA_PORT", "9092"),
		KafkaTranscodingTopic:     getEnvWithDefault("APPLICATION_KAFKA_TRANSCODING_TOPIC", "video-reencode-test"),
		KafkaAudioExtractionTopic: getEnvWithDefault("APPLICATION_KAFKA_AUDIO_EXTRACTION_TOPIC", "video-audio-extraction-test"),
		BlobConnectString:         os.Getenv("AZURE_BLOB_CONNECTION_STRING"),
		VideoContainerName:        os.Getenv("VIDEO_CONTAINER_NAME"),
		TranscodingMaxRetry:       3,
		EventHubNamespace:         os.Getenv("KAFKA_EVENTHUB_NAMESPACE"),
		EventHubName:              os.Getenv("KAFKA_EVENTHUB_NAME"),
		EventHubConnectionString:  os.Getenv("KAFKA_EVENTHUB_CONNECTION_STRING"),
		Username:                  os.Getenv("KAFKA_USERNAME"),
		AudioContainerName:        os.Getenv("AUDIO_CONTAINER_NAME"),
		AudioExtractionMaxRetry:   3,
		AudioSamplingRate:         getEnvWithDefault("AUDIO_SAMPLING_RATE", "8000"),
		AudioBitRate:              getEnvWithDefault("AUDIO_BIT_RATE", "16"),
		KafkaBootstrapServers:     os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),
	}
}

func GetInstance() *Variables {
	if instance == nil {
		once.Do(loadValues)
	}
	return instance
}
