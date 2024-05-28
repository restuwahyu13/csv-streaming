package main

import (
	"context"

	"restuwahyu13/csv-stream/helpers"
	"restuwahyu13/csv-stream/packages"
)

type service struct {
	ctx    context.Context
	broker packages.Ikafka
}

func NewService(ctx context.Context, broker packages.Ikafka) IService {
	return &service{ctx: ctx, broker: broker}
}

func (s *service) UploadCsvFile(pool bool, body []byte) error {
	var (
		publisherTopicName string          = "csv.stream"
		parser             helpers.IParser = helpers.NewParser()
	)

	csv, err := parser.CsvToJson(pool, body)
	if err != nil {
		return err
	}

	if err := s.broker.Publisher(publisherTopicName, nil, &csv); err != nil {
		return err
	}

	return nil
}
