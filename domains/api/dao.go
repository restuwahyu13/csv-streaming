package main

import "net/http"

type (
	IHandler interface {
		UploadCsvFile(rw http.ResponseWriter, r *http.Request)
	}

	IService interface {
		UploadCsvFile(isPool bool, body []byte) error
	}
)
