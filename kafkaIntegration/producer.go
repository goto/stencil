package kafkaIntegration

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
	"log"
)

type KafkaProducer struct {
	hostName string
	producer *kafka.Producer
}

func NewKafkaProducer(hostName string) (*KafkaProducer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": hostName})
	if err != nil {
		log.Printf("failed to create producer: %w", err)
		return nil, err
	}
	return &KafkaProducer{producer: producer, hostName: hostName}, nil
}

// ProduceMessages sends a message to the specified Kafka topic
func (kp *KafkaProducer) ProduceMessages(topic string, protoMessage proto.Message) error {
	messageBytes, err := proto.Marshal(protoMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal Protobuf message: %v", err)
	}
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          messageBytes,
	}

	deliveryChan := make(chan kafka.Event)
	err = kp.producer.Produce(message, deliveryChan)
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
