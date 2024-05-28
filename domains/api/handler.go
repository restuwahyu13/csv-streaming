package main

import (
	"bytes"
	"io"
	"net/http"

	"restuwahyu13/csv-stream/helpers"
)

type handler struct {
	service IService
}

func NewHander(service IService) IHandler {
	return &handler{service: service}
}

func (h *handler) UploadCsvFile(rw http.ResponseWriter, r *http.Request) {
	var (
		parser      helpers.IParser = helpers.NewParser()
		bytesWriter *bytes.Buffer   = new(bytes.Buffer)
		pool        bool            = false
	)

	if err := r.ParseMultipartForm(1e+9); err != nil {
		helpers.ApiResponse(rw, &helpers.APIResponse{
			StatCode:   http.StatusFailedDependency,
			ErrMessage: err.Error(),
		})
		return
	}

	csvFile, _, err := r.FormFile("file")
	if err != nil {
		helpers.ApiResponse(rw, &helpers.APIResponse{
			StatCode:   http.StatusFailedDependency,
			ErrMessage: err.Error(),
		})
		return
	}

	if _, err := io.Copy(bytesWriter, csvFile); err != nil {
		helpers.ApiResponse(rw, &helpers.APIResponse{
			StatCode:   http.StatusFailedDependency,
			ErrMessage: err.Error(),
		})
		return
	}

	if contentLength, _ := parser.ToInt(r.Header.Get("content-length")); contentLength < 314572800 {
		pool = true
	}

	if err := h.service.UploadCsvFile(pool, bytesWriter.Bytes()); err != nil {
		helpers.ApiResponse(rw, &helpers.APIResponse{
			StatCode:   http.StatusFailedDependency,
			ErrMessage: err.Error(),
		})
		return
	}

	helpers.ApiResponse(rw, &helpers.APIResponse{
		StatCode:    http.StatusOK,
		StatMessage: "Uploading csv file sucessfully",
	})
}
