package config

import "github.com/gofrs/uuid"

// send values through middleware in context

type CtxConfig struct {
	userID       *uuid.UUID
	isRegistered bool
}

func NewCtxConfig(userID *uuid.UUID, isRegistered bool) CtxConfig {

	return CtxConfig{userID: userID, isRegistered: isRegistered}
}

func (c CtxConfig) GetUserID() *uuid.UUID {
	return c.userID
}

func (c CtxConfig) IsRegistered() bool {
	return c.isRegistered
}
