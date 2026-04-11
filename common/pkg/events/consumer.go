package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type ConsumerConfig struct {
	Brokers      []string
	GroupID      string
	DLQSuffix    string
	FetchMaxWait time.Duration
}

func DefaultConsumerConfig(brokers []string, groupID string) ConsumerConfig {
	return ConsumerConfig{
		Brokers:      brokers,
		GroupID:      groupID,
		DLQSuffix:    ".dlq",
		FetchMaxWait: 500 * time.Millisecond,
	}
}

type franzClient struct {
	client *kgo.Client
	logger *slog.Logger
	dlq    *kgo.Client
	router *Router
	cfg    ConsumerConfig
}

type Consumer interface {
	Start(ctx context.Context) error
	Close()
}

func NewConsumer(cfg ConsumerConfig, routeHandler *Router, logger *slog.Logger) (Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(cfg.GroupID),
		kgo.Balancers(kgo.CooperativeStickyBalancer()),
		kgo.DisableAutoCommit(),
		kgo.FetchMaxWait(cfg.FetchMaxWait),
		kgo.FetchMinBytes(1),
	)
	if err != nil {
		fmt.Println("failed to create Kafka consumer:", err)
		return nil, err
	}

	dlqClient, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	)

	if err != nil {
		fmt.Println("failed to create Kafka consumer:", err)
		return nil, err
	}

	return &franzClient{
		client: client,
		dlq:    dlqClient,
		router: routeHandler,
		logger: logger,
		cfg:    cfg,
	}, nil
}

func (c *franzClient) Start(ctx context.Context) error {
	topics := c.router.GetTopics()
	if len(topics) == 0 {
		c.logger.Info("no topics defined")
		return nil
	}
	c.client.AddConsumeTopics(topics...)
	c.logger.Info("consumer started")
	for {
		fetches := c.client.PollFetches(ctx)
		if err := fetches.Err0(); err != nil {
			if errors.Is(err, context.Canceled) {
				c.logger.Warn("context canceled, stopping consumer gracefully")
				return nil
			}
			c.logger.Warn("fetch error continue to next message", "error", err.Error())
			continue
		}

		var wg sync.WaitGroup
		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			wg.Add(1)
			partition := p
			go func() {
				defer wg.Done()
				c.handlePartition(ctx, partition)
			}()
		})

		wg.Wait()

		if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
			c.logger.Warn("failed to commit uncommitted offsets", "error", err.Error())
		}
	}
}

func (c *franzClient) sendToDlq(ctx context.Context, original *kgo.Record, handleErr error) {
	dlqTopic := original.Topic + c.cfg.DLQSuffix
	headers := make([]kgo.RecordHeader, len(original.Headers), len(original.Headers)+4)
	copy(headers, original.Headers)
	headers = append(headers,
		kgo.RecordHeader{Key: "original-topic", Value: []byte(original.Topic)},
		kgo.RecordHeader{Key: "original-partition", Value: []byte(fmt.Sprintf("%d", original.Partition))},
		kgo.RecordHeader{Key: "original-offset", Value: []byte(fmt.Sprintf("%d", original.Offset))},
		kgo.RecordHeader{Key: "error", Value: []byte(handleErr.Error())},
	)

	rec := &kgo.Record{
		Topic:   dlqTopic,
		Key:     original.Key,
		Value:   original.Value,
		Headers: headers,
	}

	results := c.dlq.ProduceSync(ctx, rec)
	if err := results.FirstErr(); err != nil {
		c.logger.Error("CRITICAL: DLQ produce failed — message may be lost",
			"dlq_topic", dlqTopic,
			"original_topic", original.Topic,
			"original_offset", original.Offset,
			"err", err,
		)
	} else {
		c.logger.Warn("message sent to DLQ",
			"dlq_topic", dlqTopic,
			"original_topic", original.Topic,
			"original_offset", original.Offset,
		)
	}
}

func (c *franzClient) Close() {
	c.logger.Info("closing consumer")
	c.client.Close()
	c.dlq.Close()
	c.logger.Info("consumer closed")
}

func (c *franzClient) handlePartition(ctx context.Context, p kgo.FetchTopicPartition) {
	p.EachRecord(func(rec *kgo.Record) {
		msg := c.toMessage(rec)
		err := c.router.Route(ctx, msg)
		if err != nil {
			c.logger.Error("message routing failed, sending to DLQ",
				"topic", rec.Topic,
				"partition", rec.Partition,
				"offset", rec.Offset,
				"key", string(rec.Key),
				"err", err,
			)
		}

		c.client.MarkCommitRecords(rec)
	})
}

func (c *franzClient) toMessage(rec *kgo.Record) Message {
	headers := make(map[string]string, len(rec.Headers))
	for _, h := range rec.Headers {
		headers[h.Key] = string(h.Value)
	}
	return Message{
		Key:     rec.Key,
		Value:   rec.Value,
		Headers: headers,
		Topic:   rec.Topic,
	}
}
