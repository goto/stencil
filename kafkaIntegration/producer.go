package kafkaIntegration

import (
	"fmt"
	"github.com/cactus/go-statsd-client/v5/statsd"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
	"log"
	"time"
)

type KafkaProducer struct {
	hostName string
	producer *kafka.Producer
	metrics  *ProducerMetrics
}

func NewKafkaProducer(hostName string) (*KafkaProducer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": hostName})
	if err != nil {
		log.Printf("failed to create producer: %w", err)
		return nil, err
	}

	return &KafkaProducer{producer: producer, hostName: hostName, metrics: &ProducerMetrics{
		SuccessCount: 0,
		FailureCount: 0,
		Latency:      0,
	}}, nil
}

// ProduceMessages sends a message to the specified Kafka topic
func (kp *KafkaProducer) ProduceMessages(topic string, protoMessage proto.Message, statsdClient statsd.Statter) error {
	messageBytes, err := proto.Marshal(protoMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal Protobuf message: %v", err)
	}
	const maxRetries = 3
	const retryInterval = 1 * time.Second

	startTime := time.Now()

	for i := 1; i <= maxRetries; i++ {
		message := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          messageBytes,
		}

		deliveryChan := make(chan kafka.Event)
		err = kp.producer.Produce(message, deliveryChan)
		if err != nil {
			fmt.Printf("failed to produce message (attempt %d/%d): %v\n", i, maxRetries, err)
			time.Sleep(retryInterval)
			continue
		}

		deliveryReport := <-deliveryChan
		m := deliveryReport.(*kafka.Message)

		if m.TopicPartition.Error != nil {
			fmt.Printf("delivery failed (attempt %d/%d): %v\n", i, maxRetries, m.TopicPartition.Error)
			time.Sleep(retryInterval)
			continue
		}

		endTime := time.Now()
		latency := endTime.Sub(startTime)
		kp.metrics.Latency = latency

		kp.metrics.SuccessCount++

		err := statsdClient.Gauge("kafka_integration.latency", latency.Milliseconds(), 1)
		if err != nil {
			return err
		}

		err = statsdClient.Inc("kafka_integration.success_count", int64(kp.metrics.SuccessCount), 1)
		if err != nil {
			return err
		}
		return nil
	}
	kp.metrics.FailureCount++
	err = statsdClient.Inc("kafka_integration.success_count", int64(kp.metrics.FailureCount), 1)
	if err != nil {
		return err
	}
	return fmt.Errorf("failed to produce message after %d retries", maxRetries)
}
