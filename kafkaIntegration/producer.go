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

const (
	LatencyMetric      = "kafkaIntegration.Latency"
	SuccessCountMetric = "kafkaIntegration.SuccessCount"
	FailureCountMetric = "kafkaIntegration.FailureCount"
)

func NewKafkaProducer(hostName string) (*KafkaProducer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": hostName})
	if err != nil {
		log.Printf(err.Error())
		return nil, err
	}

	return &KafkaProducer{producer: producer, hostName: hostName, metrics: &ProducerMetrics{
		SuccessCount: 0,
		FailureCount: 0,
		Latency:      0,
	}}, nil
}

func (kp *KafkaProducer) PushMessagesWithRetries(topic string, protoMessage proto.Message, statsdClient statsd.Statter, retries int, retryInterval time.Duration) error {
	messageBytes, err := proto.Marshal(protoMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal Protobuf message: %v", err)
	}
	startTime := time.Now()
	for i := 1; i <= retries; i++ {
		err := kp.PushMessages(messageBytes, topic)
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}
		endTime := time.Now()
		latency := endTime.Sub(startTime)
		kp.metrics.Latency = latency

		kp.metrics.SuccessCount++

		err = statsdClient.Gauge(LatencyMetric, latency.Milliseconds(), 1)
		if err != nil {
			log.Printf(err.Error())
		}

		err = statsdClient.Inc(SuccessCountMetric, int64(kp.metrics.SuccessCount), 1)
		if err != nil {
			log.Printf(err.Error())
		}
		return nil
	}
	kp.metrics.FailureCount++
	err = statsdClient.Inc(FailureCountMetric, int64(kp.metrics.FailureCount), 1)
	if err != nil {
		log.Printf(err.Error())
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
		log.Printf(err.Error())
		return err
	}

	deliveryReport := <-deliveryChan
	m := deliveryReport.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		log.Printf(err.Error())
		return err
	}
	return nil
}
