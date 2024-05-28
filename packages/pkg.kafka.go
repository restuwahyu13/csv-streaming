package packages

import (
	"context"
	"encoding/json"
	"os"

	"github.com/panjf2000/ants/v2"
	kafkaBroker "github.com/segmentio/kafka-go"

	"restuwahyu13/csv-stream/helpers"
)

type (
	Ikafka interface {
		Publisher(topic string, key, value interface{}) error
		Consumer(topic, groupId string, handler func(message kafkaBroker.Message) error) (*kafkaBroker.Reader, error)
	}

	kafka struct {
		ctx     context.Context
		brokers []string
	}
)

func NewKafka(ctx context.Context, brokers []string) Ikafka {
	return &kafka{ctx: ctx, brokers: brokers}
}

func (h *kafka) Publisher(topic string, key, value interface{}) error {
	broker := kafkaBroker.Writer{
		Addr:                   kafkaBroker.TCP(h.brokers...),
		Topic:                  topic,
		Compression:            kafkaBroker.Snappy,
		RequiredAcks:           kafkaBroker.RequireAll,
		AllowAutoTopicCreation: true,
		BatchBytes:             1e+9,
		MaxAttempts:            10,
		BatchSize:              0,
		ErrorLogger: kafkaBroker.LoggerFunc(func(s string, i ...interface{}) {
			Logrus("error", s)
		}),
	}

	pool, err := ants.NewPoolWithFunc(1000, func(data interface{}) {
		body := data.([]byte)

		msg := kafkaBroker.Message{Value: body}
		if key != nil {
			msg = kafkaBroker.Message{Key: []byte(key.(string)), Value: body}
		}

		if err := broker.WriteMessages(h.ctx, msg); err != nil {
			Logrus("error", err)
			return
		}
	},
		ants.WithPreAlloc(true),
		ants.WithNonblocking(true),
	)

	if err != nil {
		return err
	}

	bodyByte, err := json.Marshal(value)
	if err != nil {
		return err
	}

	defer pool.Release()
	if err := pool.Invoke(bodyByte); err != nil {
		return err
	}

	return nil
}

func (h *kafka) Consumer(topic, groupId string, handler func(message kafkaBroker.Message) error) (*kafkaBroker.Reader, error) {
	parser := helpers.NewParser()

	gorutinePoolSize, err := parser.ToInt(os.Getenv("GORUTINE_POOL_SIZE"))
	if err != nil {
		return nil, err
	}

	pool, err := ants.NewPoolWithFunc(gorutinePoolSize, func(data interface{}) {
		reader := data.(*kafkaBroker.Reader)

		for {
			message, err := reader.ReadMessage(h.ctx)
			if err != nil {
				Logrus("error", err)
				return
			}

			if err := handler(message); err != nil {
				Logrus("error", err)
				return
			}
		}
	},
		ants.WithPreAlloc(true),
		ants.WithNonblocking(true),
	)

	if err != nil {
		return nil, err
	}

	reader := kafkaBroker.NewReader(kafkaBroker.ReaderConfig{
		Brokers:                h.brokers,
		Topic:                  topic,
		GroupID:                groupId,
		OffsetOutOfRangeError:  true,
		MaxBytes:               1e+9,
		MaxAttempts:            10,
		PartitionWatchInterval: 3,
		JoinGroupBackoff:       3,
		GroupBalancers:         []kafkaBroker.GroupBalancer{kafkaBroker.RoundRobinGroupBalancer{}},
		ErrorLogger: kafkaBroker.LoggerFunc(func(s string, i ...interface{}) {
			Logrus("error", s)
		}),
	})

	defer pool.Release()
	if err := pool.Invoke(reader); err != nil {
		return nil, err
	}

	return reader, nil
}
