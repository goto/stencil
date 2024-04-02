package kafkaIntegration

import (
	"log"
	"time"

	"github.com/cactus/go-statsd-client/v5/statsd"
)

type MetricsCollector struct {
	StatsdClient statsd.Statter
}

type ProducerMetrics struct {
	SuccessCount int
	FailureCount int
	Latency      time.Duration
}

func NewMetricsCollector(statsdAddr string) (*MetricsCollector, error) {

	statsdClient, err := statsd.NewClient(statsdAddr, "")
	if err != nil {
		log.Printf("failed to create statsd client: %v", err)
		return nil, err
	}

	return &MetricsCollector{
		StatsdClient: statsdClient,
	}, nil
}

func (mc *MetricsCollector) MonitorProducerMetrics(kp *KafkaProducer, producerMetrics *ProducerMetrics) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		producerMetrics.SuccessCount = kp.metrics.SuccessCount
		producerMetrics.FailureCount = kp.metrics.FailureCount

		producerMetrics.Latency = kp.metrics.Latency
	}

}
