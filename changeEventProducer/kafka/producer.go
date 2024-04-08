package kafka

import (
	"fmt"
	"github.com/cactus/go-statsd-client/v5/statsd"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
	"log"
	"time"
)

type KafkaProducer struct {
	hostName     string
	producer     *kafka.Producer
	statsdClient statsd.Statter
}

const (
	SuccessCountMetric = "kafka.SuccessCount"
	FailureCountMetric = "kafka.FailureCount"
)

func NewKafkaProducer(hostName string, statsdClient statsd.Statter) (*KafkaProducer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": hostName})
	if err != nil {
		log.Printf("Failed to initialise Kafka Producer - %s", err.Error())
		return nil, err
	}

	return &KafkaProducer{producer: producer, hostName: hostName, statsdClient: statsdClient}, nil
}

func (kp *KafkaProducer) PushMessagesWithRetries(topic string, protoMessage proto.Message, retries int, retryInterval time.Duration) error {
	messageBytes, err := proto.Marshal(protoMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal Protobuf message: %v", err)
	}

	for i := 0; i < retries; i++ {
		err := kp.PushMessages(messageBytes, topic)
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}

		err = kp.statsdClient.Inc(SuccessCountMetric, 1, 1)
		if err != nil {
			log.Printf("Failed to increase Success metric - %s", err.Error())
		}
		return nil
	}
	err = kp.statsdClient.Inc(FailureCountMetric, 1, 1)
	if err != nil {
		log.Printf("Failed to increase Failure metric - %s", err.Error())
	}
	return fmt.Errorf("failed to produce message after %d retries", retries)
}

func (kp *KafkaProducer) PushMessages(messageBytes []byte, topic string) error {

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          messageBytes,
	}

	deliveryChan := make(chan kafka.Event)
	err := kp.producer.Produce(message, deliveryChan)
	if err != nil {
		log.Printf("Error in producing messages- %s", err.Error())
		return err
	}

	deliveryReport := <-deliveryChan
	if m, ok := deliveryReport.(*kafka.Message); ok && m.TopicPartition.Error != nil {
		log.Printf("Error in topic partitioning- %s", err.Error())
		return err
	}

	return nil
}
