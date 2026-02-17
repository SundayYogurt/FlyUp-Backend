package queue

import (
	"context"
	"log"

	"github.com/SundayYogurt/user_service/internal/interfaces"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	Reader      *kafka.Reader
	Handler     interfaces.ConsumerHandler
	ServiceName string
}

func NewKafkaConsumer(broker, topic, groupID string, handler interfaces.ConsumerHandler) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, //10KB
		MaxBytes: 10e6, //10MB
	})

	return &KafkaConsumer{
		Reader:      reader,
		Handler:     handler,
		ServiceName: "User Service",
	}
}

func (kc *KafkaConsumer) Listen() {
	// Listen for messages continuously
	for {
		msg, err := kc.Reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error on reading message: %s\n", err)
			continue
		}

		log.Printf("Received message: %s\n", string(msg.Value))

		if err := kc.Handler.HandleMessage(string(msg.Value)); err != nil {
			log.Printf("Error on processing message on handler: %s\n", err)
		}
	}
}
