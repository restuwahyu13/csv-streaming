package configs

type Environtment struct {
	PORT               string `json:"PORT" env:"PORT" envDefault:"3000"`
	DSN                string `json:"DSN" env:"DSN" envDefault:"postgres://admin:admin@localhost:5432/postgres?sslmode=disable"`
	BSN                string `json:"BSN" env:"BSN" envDefault:"localhost:9092"`
	GORUTINE_POOL_SIZE string `json:"GORUTINE_POOL_SIZE" env:"GORUTINE_POOL_SIZE" envDefault:"1000"`
	CHUNK_CSV_SIZE     string `json:"CHUNK_CSV_SIZE" env:"CHUNK_CSV_SIZE" envDefault:"100000"`
}
