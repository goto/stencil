package kafkaIntegration

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
	"log"
)

type Producer struct {
	hostName string
	producer *kafka.Producer
}

func NewProducer(hostName string) *Producer {
	return &Producer{
		hostName: hostName,
	}
}

func (p *Producer) Initialize() error {
	var err error
	configMap := &kafka.ConfigMap{
		"bootstrap.servers": p.hostName,
	}
	p.producer, err = kafka.NewProducer(configMap)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	return err
}

func produce(producer *kafka.Producer, topic string, protoMessage proto.Message) error {
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

func (p *Producer) ProduceMessage(topic string, protoMessage proto.Message) error {
	if err := produce(p.producer, topic, protoMessage); err != nil {
		fmt.Printf("Error producing message: %v\n", err)
	}
	return nil
}
