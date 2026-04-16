package events

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer interface {
	Publish(ctx context.Context, topic string, msg Message) error
	PublishBatch(ctx context.Context, topic string, msgs []Message) error
	Close()
}

type ProducerConfig struct {
	Brokers        []string
	MaxRetries     int
	RetryBackoff   time.Duration
	MaxBatchBytes  int32
	ProduceTimeout time.Duration
}

func GetDefaultConfig(brokers []string) ProducerConfig {
	return ProducerConfig{
		Brokers:        brokers,
		MaxBatchBytes:  1_000_000,
		MaxRetries:     3, // ✅ Reduced for dev
		RetryBackoff:   100 * time.Millisecond,
		ProduceTimeout: 10 * time.Second,
	}
}

type franzKafka struct {
	client *kgo.Client
}

func NewProducer(cfg ProducerConfig) (Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ProducerBatchMaxBytes(cfg.MaxBatchBytes),

		//Use LeaderAck for single-node dev
		kgo.RequiredAcks(kgo.LeaderAck()),
		kgo.DisableIdempotentWrite(),

		////Multi Node
		//kgo.RequiredAcks(kgo.AllISRAcks()),
		////kgo.DisableIdempotentWrite(),

		kgo.ProduceRequestTimeout(cfg.ProduceTimeout),
		kgo.RetryBackoffFn(func(attempt int) time.Duration {
			return cfg.RetryBackoff * time.Duration(1<<uint(attempt-1))
		}),
		kgo.RecordRetries(cfg.MaxRetries),
		kgo.MetadataMaxAge(30*time.Second),

		// Logger
		//kgo.WithLogger(kgo.BasicLogger(log.Writer(), kgo.LogLevelDebug, nil)),
	)

	if err != nil {
		fmt.Println("failed to create Kafka producer:", err)
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	log.Println("✅ Kafka producer connected successfully")

	return &franzKafka{client: client}, nil
}

func (r *franzKafka) Publish(ctx context.Context, topic string, msg Message) error {
	kafkaCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rec := r.toRecord(topic, msg)

	result := r.client.ProduceSync(kafkaCtx, rec)
	if err := result.FirstErr(); err != nil {
		fmt.Println("failed to publish message:", err)
		return err
	}
	return nil
}

func (r *franzKafka) PublishBatch(ctx context.Context, topic string, msgs []Message) error {
	kafkaCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if len(msgs) == 0 {
		return nil
	}

	records := make([]*kgo.Record, len(msgs))
	for i, m := range msgs {
		records[i] = r.toRecord(topic, m)
	}

	result := r.client.ProduceSync(kafkaCtx, records...)
	if err := result.FirstErr(); err != nil {
		fmt.Println("failed to publish batch messages:", err)
		return err
	}
	return nil
}

func (r *franzKafka) Close() {
	log.Println("🔌 Closing Kafka producer...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r.client.Flush(ctx)
	r.client.Close()
	log.Println("✅ Kafka producer closed")
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
