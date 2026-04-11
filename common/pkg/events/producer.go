package events

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer interface {
	Publish(ctx context.Context, topic string, msg Message) error
	PublishBatch(ctx context.Context, topic string, msgs []Message) error
	Close()
}

type ProducerConfig struct {
	Brokers       []string
	MaxRetries    int
	RetryBackoff  time.Duration
	MaxBatchBytes int32
}

func GerDefaultConfig(brokers []string) ProducerConfig {
	return ProducerConfig{
		Brokers:       brokers,
		MaxBatchBytes: 1_000_000,
		MaxRetries:    5,
		RetryBackoff:  250 * time.Millisecond,
	}
}

type franzKafka struct {
	client *kgo.Client
}

func NewProducer(cfg ProducerConfig) (Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ProducerBatchMaxBytes(cfg.MaxBatchBytes),
		kgo.RetryBackoffFn(func(retries int) time.Duration {
			return cfg.RetryBackoff
		}),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	)

	if err != nil {
		fmt.Println("failed to create Kafka producer:", err)
		return nil, err
	}
	return &franzKafka{client: client}, nil
}

func (r *franzKafka) Publish(ctx context.Context, topic string, msg Message) error {
	rec := r.toRecord(topic, msg)

	result := r.client.ProduceSync(ctx, rec)
	if err := result.FirstErr(); err != nil {
		fmt.Println("failed to publish message:", err)
		return err
	}
	return nil
}

func (r *franzKafka) PublishBatch(ctx context.Context, topic string, msgs []Message) error {
	if len(msgs) == 0 {
		return nil
	}

	records := make([]*kgo.Record, len(msgs))
	for i, m := range msgs {
		records[i] = r.toRecord(topic, m)
	}

	result := r.client.ProduceSync(ctx, records...)
	if err := result.FirstErr(); err != nil {
		fmt.Println("failed to publish batch messages:", err)
		return err
	}
	return nil
}

func (r *franzKafka) Close() {
	r.client.Close()
}

func (r *franzKafka) toRecord(topic string, msg Message) *kgo.Record {
	rec := &kgo.Record{
		Topic: topic,
		Key:   msg.Key,
		Value: msg.Value,
	}

	for k, v := range msg.Headers {
		rec.Headers = append(rec.Headers, kgo.RecordHeader{
			Key:   k,
			Value: []byte(v),
		})
	}
	return rec
}
