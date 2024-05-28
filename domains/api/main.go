package main

import (
	"context"

	"github.com/go-chi/chi/v5"

	"restuwahyu13/csv-stream/configs"
	"restuwahyu13/csv-stream/packages"
)

func main() {
	var (
		env *configs.Environtment = new(configs.Environtment)
		ctx context.Context       = context.Background()
	)

	err := packages.ViperRead(".env", env)
	if err != nil {
		packages.Logrus("error", err)
		return
	}

	broker := packages.NewKafka(ctx, []string{env.BSN})

	service := NewService(ctx, broker)
	handler := NewHander(service)

	router := chi.NewMux()
	router.Post("/", handler.UploadCsvFile)

	err = packages.Graceful(func() *packages.GracefulConfig {
		return &packages.GracefulConfig{Handler: router, Port: env.PORT}
	})

	if err != nil {
		packages.Logrus("fatal", "Server is not running: %s", err.Error())
	}

	packages.Logrus("info", "Server is running on port: %s", env.PORT)
}
