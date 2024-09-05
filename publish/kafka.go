package publish

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"strconv"

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
			TLS: &tls.Config{
				InsecureSkipVerify: false,
			}, // Ensure TLS is configured for Azure Event Hubs
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

	if eventHubNamespace == "" { // Only attempt topic creation in local Kafka setup
		err := createTopicIfNotExists(addr, topic)
		if err != nil {
			log.Printf("Failed to create topic %s: %s", topic, err)
			return err
		}
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

func createTopicIfNotExists(addr string, topic string) error {
	// Connect to Kafka broker
	conn, err := kafka.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Get controller information
	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	// Connect to controller to create topic
	connController, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer connController.Close()

	// Define topic configuration
	config := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	// Create topic
	err = connController.CreateTopics(config)
	if err != nil {
		return err
	}

	log.Printf("Created topic %s", topic)
	return nil
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
