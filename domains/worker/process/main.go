package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/segmentio/kafka-go"

	"restuwahyu13/csv-stream/configs"
	"restuwahyu13/csv-stream/helpers"
	"restuwahyu13/csv-stream/models"
	"restuwahyu13/csv-stream/packages"
)

/**
=======================================================
=	WORKER MAIN TERITORY
=======================================================
**/

type worker struct {
	env     *configs.Environtment
	broker  packages.Ikafka
	handler IHandler
}

func main() {
	var (
		env    *configs.Environtment = new(configs.Environtment)
		ctx    context.Context       = context.Background()
		broker packages.Ikafka       = packages.NewKafka(ctx, []string{env.BSN})
	)

	err := packages.ViperRead(".env", env)
	if err != nil {
		packages.Logrus("error", err)
		return
	}

	db, err := packages.Database(env.DSN)
	if err != nil {
		packages.Logrus("fatal", err)
		return
	}

	service := NewService(ctx, db, broker)
	handler := NewHandler(service)

	worker := NewWorker(env, broker, handler)
	worker.Listener()
}

/**
=======================================================
=	WORKER HANDLER TERITORY
=======================================================
**/

func NewWorker(env *configs.Environtment, broker packages.Ikafka, handler IHandler) IWorker {
	return &worker{env: env, broker: broker, handler: handler}
}

func (w *worker) Pool() (*ants.PoolWithFunc, error) {
	parser := helpers.NewParser()

	gorutinePoolSize, err := parser.ToInt(w.env.GORUTINE_POOL_SIZE)
	if err != nil {
		return nil, err
	}

	pool, err := ants.NewPoolWithFunc(gorutinePoolSize, func(data interface{}) {
		var (
			message kafka.Message = data.(kafka.Message)
			users   []models.User = []models.User{}
		)

		if err := parser.Unmarshal(message.Value, &users); err != nil {
			packages.Logrus("error", err)
			return
		}

		w.handler.UpsertUser(users)
	},
		ants.WithPreAlloc(true),
		ants.WithNonblocking(true),
	)

	if err != nil {
		return pool, err
	}

	return pool, nil
}

func (w *worker) Consumer(pool *ants.PoolWithFunc) (*kafka.Reader, error) {
	var (
		consumerTopicName string = "csv.process"
		consumerGroupName string = "csv-process-group"
	)

	defer pool.Release()
	res, err := w.broker.Consumer(consumerTopicName, consumerGroupName, func(message kafka.Message) error {
		if err := pool.Invoke(message); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return res, err
	}

	return res, nil
}

func (w *worker) Listener() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGALRM, syscall.SIGABRT, syscall.SIGUSR1)

	pool, err := w.Pool()
	if err != nil {
		packages.Logrus("error", err)
		return
	}

	consumer, err := w.Consumer(pool)
	if err != nil {
		packages.Logrus("error", err)
		return
	}

	defer consumer.Close()
	for {
		select {
		case <-signalChan:
			os.Exit(0)

		default:
			time.Sleep(time.Second * 3)
			packages.Logrus("info", "Worker Process Is Running ....")
		}
	}
}
