package queue

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker, topic string) *Producer {
	// ถ้า config ว่าง ให้ไม่พังระบบ
	if broker == "" || topic == "" {
		log.Println("Kafka broker/topic empty - producer disabled")
		return &Producer{writer: nil}
	}

	// create topic best-effort (fail ก็แค่ log)
	if err := createTopic(broker, topic); err != nil {
		log.Printf("Error creating topic: %v\n", err)
	}

	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(broker),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
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

func createTopic(broker, topic string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return err
	}

	for _, p := range partitions {
		if p.Topic == topic {
			return nil
		}
	}

	return conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
}
