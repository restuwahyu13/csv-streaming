.PHONY: worker
worker:
ifdef type
	cd domains; nodemon -V -e .go -w . -x go run --race -V --signal SIGTERM  ./worker/${type}
endif

.PHONY: api
api:
	cd domains; nodemon -V -e .go -w . -x go run --race -V --signal SIGTERM ./api