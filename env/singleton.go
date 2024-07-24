package env

import (
	"os"
	"sync"
)

type Variables struct {
	Debug                    bool
	KafkaHost                string
	KafkaPort                string
	KafkaTopic               string
	BlobConnectString        string
	VideoContainerName       string
	TranscodingMaxRetry      int
	EventHubNamespace        string
	EventHubName             string
	EventHubConnectionString string
	Username                 string
}

var instance *Variables
var once sync.Once

func loadValues() {
	instance = &Variables{
		Debug:                    os.Getenv("APPLICATION_DEBUG") == "true",
		KafkaHost:                os.Getenv("APPLICATION_KAFKA_HOST"),
		KafkaPort:                os.Getenv("APPLICATION_KAFKA_PORT"),
		KafkaTopic:               os.Getenv("APPLICATION_KAFKA_TOPIC"),
		BlobConnectString:        os.Getenv("AZURE_BLOB_CONNECTION_STRING"),
		VideoContainerName:       os.Getenv("VIDEO_CONTAINER_NAME"),
		TranscodingMaxRetry:      3,
		EventHubNamespace:        os.Getenv("KAFKA_EVENTHUB_NAMESPACE"),
		EventHubName:             os.Getenv("KAFKA_EVENTHUB_NAME"),
		EventHubConnectionString: os.Getenv("KAFKA_EVENTHUB_CONNECTION_STRING"),
		Username:                 os.Getenv("KAFKA_USERNAME"),
	}
}

func GetInstance() *Variables {
	if instance == nil {
		once.Do(loadValues)
	}
	return instance
}
