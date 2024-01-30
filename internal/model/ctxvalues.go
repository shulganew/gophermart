package model

import "github.com/gofrs/uuid"

// send pass to midleware
type CtxPassKey struct{}

// send values through middleware in context

type MiddlwDTO struct {
	userID       *uuid.UUID
	isRegistered bool
}

func NewMiddlwDTO(userID *uuid.UUID, isRegistered bool) MiddlwDTO {

	return MiddlwDTO{userID: userID, isRegistered: isRegistered}
}

func (c MiddlwDTO) GetUserID() *uuid.UUID {
	return c.userID
}

func (c MiddlwDTO) IsRegistered() bool {
	return c.isRegistered
}
