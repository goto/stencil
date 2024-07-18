package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

var (
	kafkaSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_success_count",
		Help: "Success count of the successfully processed kafka messages",
	})
	kafkaFailure = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_messages_failure_count",
		Help: "Failure count of the kafka messages",
	}, []string{"cause"})
)

type Writer struct {
	kafkaWriter *kafka.Writer
}

func NewWriter(kafkaBrokerUrl string, timeout int, retries int) *Writer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(kafkaBrokerUrl),
		WriteTimeout: time.Second * time.Duration(timeout),
		MaxAttempts:  retries,
	}

	return &Writer{kafkaWriter: writer}
}

func (w *Writer) Close() error {
	return w.kafkaWriter.Close()
}

func (w *Writer) Write(topic string, protoMessage proto.Message) error {
	messageBytes, err := proto.Marshal(protoMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal Protobuf message: %s", err.Error())
	}

	kafkaMessage := kafka.Message{
		Topic: topic,
		Value: messageBytes,
	}
	return w.send(kafkaMessage)
}

func (w *Writer) send(message kafka.Message) error {
	err := w.kafkaWriter.WriteMessages(context.Background(), message)
	if err != nil {
		kafkaFailure.WithLabelValues("kafka_message_produce_error" + " " + err.Error()).Inc()
		return err
	}
	kafkaSuccess.Inc()
	return nil
}
