package model

import "github.com/gofrs/uuid"

type User struct {
	uuid     uuid.UUID
	login    string
	password string
}
