package queue

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker, topic, username, password string) *Producer {

	mechanism := plain.Mechanism{
		Username: username,
		Password: password,
	}

	transport := &kafka.Transport{
		SASL: mechanism,
		TLS:  &tls.Config{},
	}

	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(broker),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
			Async:        false,
			Transport:    transport,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (p *Producer) PublishMessage(key, value []byte) error {
	// ถ้า kafka ไม่พร้อม ให้ skip (ไม่ทำให้ register ล้ม)
	if p == nil || p.writer == nil {
		log.Println("Kafka producer not ready - skip publish")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	})
}

//func createTopic(broker, topic string) error {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	conn, err := kafka.DialContext(ctx, "tcp", broker)
//	if err != nil {
//		return err
//	}
//	defer conn.Close()
//
//	partitions, err := conn.ReadPartitions()
//	if err != nil {
//		return err
//	}
//
//	for _, p := range partitions {
//		if p.Topic == topic {
//			return nil
//		}
//	}
//
//	return conn.CreateTopics(kafka.TopicConfig{
//		Topic:             topic,
//		NumPartitions:     1,
//		ReplicationFactor: 1,
//	})
//}
