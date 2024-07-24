package publish

import (
	"context"
	"crypto/tls"
	"log"

	"cognitube.com/transcoding-service/env"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type Publisher interface {
	Publish(topic string, message []byte) error
}

type KafkaPublisher struct {
	Url string
}

func (p *KafkaPublisher) PublishProd(topic string, message []byte) error {
	eventHubNamespace := env.GetInstance().EventHubNamespace
	connectionString := env.GetInstance().EventHubConnectionString
	username := env.GetInstance().Username

	// Set up SASL configuration
	mechanism := plain.Mechanism{
		Username: username,
		Password: connectionString,
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP(eventHubNamespace + ".servicebus.windows.net:9093"),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		Transport: &kafka.Transport{
			SASL: mechanism,
			TLS:  &tls.Config{},
		},
	}
	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: message,
		},
	)

	if err != nil {
		log.Println(err.Error())
	}

	return err
}

func (p *KafkaPublisher) Publish(topic string, message []byte) error {
	log.Println("Try publish message to Kafka: " + string(message))

	if env.GetInstance().Debug {
		return p.PublishProd(topic, message)
	}

	w := &kafka.Writer{
		Addr:  kafka.TCP(p.Url),
		Topic: topic,
	}
	defer w.Close()
	err := w.WriteMessages(context.Background(), kafka.Message{
		Value: message,
	})

	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func NewKafkaPublisher(host, port string) Publisher {
	return &KafkaPublisher{host + ":" + port}
}
