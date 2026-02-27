package queue

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/SundayYogurt/FlyUp-Backend/mail-svc/internal/interfaces"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type KafkaConsumer struct {
	Reader      *kafka.Reader
	Handler     interfaces.ConsumerHandler
	ServiceName string
}

func NewKafkaConsumer(broker, topic, groupID, username, password string, handler interfaces.ConsumerHandler) *KafkaConsumer {
	mech := plain.Mechanism{
		Username: username,
		Password: password,
	}

	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		TLS:           &tls.Config{}, // ✅ TLS
		SASLMechanism: mech,          // ✅ SASL/PLAIN
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
		Dialer:   dialer, // ✅ สำคัญมาก
	})

	return &KafkaConsumer{
		Reader:      reader,
		Handler:     handler,
		ServiceName: "Mail Service",
	}
}

func (kc *KafkaConsumer) Listen() {
	for {
		msg, err := kc.Reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("[%s] read error: %v\n", kc.ServiceName, err)
			continue
		}

		log.Printf("[%s] received: %s\n", kc.ServiceName, string(msg.Value))

		if err := kc.Handler.HandleMessage(string(msg.Value)); err != nil {
			log.Printf("[%s] handler error: %v\n", kc.ServiceName, err)
		}
	}
}
