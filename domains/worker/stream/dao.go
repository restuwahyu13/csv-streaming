package main

import (
	"github.com/panjf2000/ants/v2"
	"github.com/segmentio/kafka-go"
)

type (
	IWorker interface {
		Pool() (*ants.PoolWithFunc, error)
		Consumer(pool *ants.PoolWithFunc) (*kafka.Reader, error)
		Listener()
	}
)
