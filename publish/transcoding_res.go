package publish

import (
	"encoding/json"

	"cognitube.com/transcoding-service/env"
	"cognitube.com/transcoding-service/result"
)

type TranscodingPublisher interface {
	PublishTranscodingResult(result *result.Result) error
}
type KafkaTranscodingPublisher struct {
	KafkaPublisher
}

func (p *KafkaTranscodingPublisher) PublishTranscodingResult(result *result.Result) error {
	msg, _ := json.Marshal(result)
	topic := env.GetInstance().KafkaTopic
	return p.Publish(topic, msg)
}

func NewKafkaTranscodingPublisher() TranscodingPublisher {
	host := env.GetInstance().KafkaHost
	port := env.GetInstance().KafkaPort
	return &KafkaTranscodingPublisher{
		KafkaPublisher: KafkaPublisher{
			Url: host + ":" + port,
		},
	}
}
