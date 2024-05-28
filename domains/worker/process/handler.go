package main

import (
	"restuwahyu13/csv-stream/models"
	"restuwahyu13/csv-stream/packages"
)

type handler struct {
	service IService
}

func NewHandler(service IService) IHandler {
	return &handler{service: service}
}

func (h *handler) UpsertUser(req []models.User) {
	if err := h.service.UpsertUser(req); err != nil {
		packages.Logrus("error", err)
		return
	}
}
