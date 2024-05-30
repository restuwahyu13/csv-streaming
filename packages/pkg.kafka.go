package packages

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/panjf2000/ants/v2"
	kafkaBroker "github.com/segmentio/kafka-go"

	"restuwahyu13/csv-stream/helpers"
)

const (
	TCP = "tcp"
)

type (
	Ikafka interface {
		Publisher(topic string, key, value interface{}) error
		Consumer(topic, groupId string, handler func(message kafkaBroker.Message) error) (*kafkaBroker.Reader, error)
		findPartitionAndOffset(protocol, address, topic string) ([]map[int]int64, error)
		consumerGroup(protocol string, topic, groupId string) error
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
	parser := helpers.NewParser()

	gorutinePoolSize, err := parser.ToInt(os.Getenv("GORUTINE_POOL_SIZE"))
	if err != nil {
		return err
	}

	broker := kafkaBroker.Writer{
		Addr:                   kafkaBroker.TCP(h.brokers...),
		Topic:                  topic,
		Compression:            kafkaBroker.Snappy,
		RequiredAcks:           kafkaBroker.RequireAll,
		AllowAutoTopicCreation: true,
		BatchBytes:             1e+9,
		BatchSize:              1000,
		MaxAttempts:            10,
		Balancer:               &kafkaBroker.LeastBytes{},
		ErrorLogger: kafkaBroker.LoggerFunc(func(msg string, args ...interface{}) {
			Logrus("error", msg)
		}),
	}

	pool, err := ants.NewPoolWithFunc(gorutinePoolSize, func(data interface{}) {
		uniqeKey := uuid.NewString()
		body := data.([]byte)

		msg := kafkaBroker.Message{Key: []byte(uniqeKey), Value: body}
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
			message, err := reader.FetchMessage(h.ctx)
			if err != nil {
				Logrus("error", err)
				return
			}

			if err := handler(message); err != nil {
				Logrus("error", err)
				return
			}

			if err := reader.CommitMessages(h.ctx, message); err != nil {
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
		MaxWait:                time.Duration(time.Second * 5),
		PartitionWatchInterval: time.Duration(time.Second * 3),
		JoinGroupBackoff:       time.Duration(time.Second * 3),
		ErrorLogger: kafkaBroker.LoggerFunc(func(msg string, args ...interface{}) {
			Logrus("error", msg)
		}),
	})

	defer pool.Release()
	if err := pool.Invoke(reader); err != nil {
		return nil, err
	}

	if err := h.consumerGroup(TCP, topic, groupId); err != nil {
		return nil, err
	}

	return reader, nil
}

func (h *kafka) findPartitionAndOffset(protocol, address, topic string) ([]map[int]int64, error) {
	con, err := kafkaBroker.Dial(protocol, address)
	if err != nil {
		return nil, err
	}

	defer con.Close()
	partitions, err := con.ReadPartitions(topic)
	if err != nil {
		return nil, err
	}

	partitionAndOffset := make(map[int]int64)
	partitionsAndOffsets := []map[int]int64{}

	for _, partition := range partitions {
		conLeader, err := kafkaBroker.DialLeader(h.ctx, protocol, net.JoinHostPort(partition.Leader.Host, strconv.Itoa(partition.Leader.Port)), partition.Topic, partition.ID)
		if err != nil {
			return nil, err
		}

		defer conLeader.Close()
		now := time.Now().Add(-time.Second * 60)

		offset, err := conLeader.ReadOffset(now)
		if err != nil {
			return nil, err
		}

		partitionAndOffset[partition.ID] = offset
		partitionsAndOffsets = append(partitionsAndOffsets, partitionAndOffset)
	}

	return partitionsAndOffsets, nil
}

func (h *kafka) consumerGroup(protocol string, topic, groupId string) error {
	for _, address := range h.brokers {
		partitionsAndOffsets, err := h.findPartitionAndOffset(protocol, address, topic)
		if err != nil {
			return err
		}

		for i, partition := range partitionsAndOffsets {
			kafkaBroker.NewConsumerGroup(kafkaBroker.ConsumerGroupConfig{
				Topics:                 []string{topic},
				ID:                     groupId,
				StartOffset:            partition[i],
				RetentionTime:          time.Duration(time.Hour * 24),
				RebalanceTimeout:       time.Duration(time.Second * 60),
				Timeout:                time.Duration(time.Second * 3),
				PartitionWatchInterval: time.Duration(time.Second * 3),
				JoinGroupBackoff:       time.Duration(time.Second * 3),
				ErrorLogger: kafkaBroker.LoggerFunc(func(msg string, args ...interface{}) {
					Logrus("error", msg)
				}),
			})
		}
	}

	return nil
}
