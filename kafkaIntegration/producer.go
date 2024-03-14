package kafkaIntegration

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
)

// ProduceMessage sends a message to the specified Kafka topic
func produceMessages(producer *kafka.Producer, topic string, protoMessage proto.Message) error {
	messageBytes, err := proto.Marshal(protoMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal Protobuf message: %v", err)
	}
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          messageBytes,
	}

	deliveryChan := make(chan kafka.Event)
	err = producer.Produce(message, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %v", err)
	}

	deliveryReport := <-deliveryChan
	m := deliveryReport.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return fmt.Errorf("delivery failed: %v", m.TopicPartition.Error)
	}

	return nil
}

func ProduceMessages(topic string, protoMessage proto.Message) error {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}
	defer producer.Close()

	if err := produceMessages(producer, topic, protoMessage); err != nil {
		fmt.Printf("Error producing message: %v\n", err)
	}

	return nil
}
