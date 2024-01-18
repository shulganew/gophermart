package model

import (
	"time"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/gofrs/uuid"
)

type Order struct {
	UserID   *uuid.UUID
	Onumber  string
	Uploaded time.Time
	Status   Status
}

func NewOrder(userID *uuid.UUID, onumber string) *Order {

	return &Order{UserID: userID, Onumber: onumber, Uploaded: time.Now(), Status: Status(0)}
}

// Check Luna namber
func (o *Order) IsValid() (isValid bool) {
	err := goluhn.Validate(o.Onumber)
	if err != nil {
		return false
	}

	return true
}
