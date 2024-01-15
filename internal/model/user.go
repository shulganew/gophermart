package model

import "github.com/gofrs/uuid"

type User struct {
	UUID     *uuid.UUID `json:"-"`
	Login    string     `json:"login"`
	Password string     `json:"password"`
}
