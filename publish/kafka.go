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
	bootstrapServers := env.GetInstance().KafkaBootstrapServers

	var addr string
	var transport kafka.Transport

	if eventHubNamespace != "" {
		addr = eventHubNamespace + ".servicebus.windows.net:9093"
		// Set up SASL configuration for Event Hubs
		mechanism := plain.Mechanism{
			Username: username,
			Password: connectionString,
		}
		transport = kafka.Transport{
			SASL: mechanism,
			TLS:  &tls.Config{}, // Ensure TLS is configured for Azure Event Hubs
		}
	} else {
		addr = bootstrapServers
		// No SASL for local Kafka
		transport = kafka.Transport{
			TLS: nil, // No TLS for local Kafka (usually not needed)
		}
	}

	writer := &kafka.Writer{
		Addr:      kafka.TCP(addr),
		Topic:     topic,
		Balancer:  &kafka.LeastBytes{},
		Transport: &transport,
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

	if !env.GetInstance().Debug {
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
