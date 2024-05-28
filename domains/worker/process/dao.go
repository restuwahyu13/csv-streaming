package main

import (
	"github.com/panjf2000/ants/v2"
	"github.com/segmentio/kafka-go"

	"restuwahyu13/csv-stream/models"
)

type (
	IHandler interface {
		UpsertUser(req []models.User)
	}

	IService interface {
		UpsertUser(req []models.User) error
	}

	IWorker interface {
		Pool() (*ants.PoolWithFunc, error)
		Consumer(pool *ants.PoolWithFunc) (*kafka.Reader, error)
		Listener()
	}
)
