package models

import "github.com/uptrace/bun"

type User struct {
	bun.BaseModel `bun:"table:user"`
	ID            string `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	FirstName     string `json:"first_name" bun:"first_name,notnull"`
	LastName      string `json:"last_name" bun:"last_name,notnull"`
	Email         string `json:"email" bun:"email,notnull"`
	Gender        string `json:"gender" bun:"gender,zeronull"`
}
