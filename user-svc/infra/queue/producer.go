package queue

import (
	"log"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker, topic string) *Producer {

	//create a new topic if it doesn't exist
	if err := createTopic(broker, topic); err != nil {
		log.Printf("Error creating topic: %v\n", err)
	}

	return &Producer{
		writer: &kafka.Writer{
			Addr:  kafka.TCP(broker),
			Topic: topic,
		},
	}
}

func (p *Producer) PublishMessage(key, value []byte) error {
	return p.writer.WriteMessages(nil, kafka.Message{
		Key:   key,
		Value: value,
	})
}

func createTopic(broker, topic string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()

	if err != nil {
		return err
	}

	for _, p := range partitions {
		log.Printf("partition: %v", p)
		if p.Topic == topic {
			return nil
		}
	}

	return conn.CreateTopics(
		kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
}
