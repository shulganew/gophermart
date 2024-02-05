package model

import (
	"encoding/json"
	"time"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

type OrderResponse struct {
	OrderNr  string  `json:"number"`
	Status   Status  `json:"status"`
	Accrual  float64 `json:"accrual,omitempty"`
	Uploaded string  `json:"uploaded_at"`
}

type Order struct {
	UserID     uuid.UUID       `db:"user_id"`
	OrderNr    string          `db:"order_number"`
	IsPreOrder bool            `db:"is_preorder"`
	Uploaded   time.Time       `db:"uploaded"`
	Status     Status          `db:"status"`
	Withdrawn  decimal.Decimal `db:"withdrawn"`
	Accrual    decimal.Decimal `db:"accrual"`
}

func NewOrder(userID uuid.UUID, orderNr string, preoreder bool, withdrawn decimal.Decimal, accrual decimal.Decimal) *Order {

	return &Order{UserID: userID, OrderNr: orderNr, IsPreOrder: preoreder, Uploaded: time.Now(), Status: Status(NEW), Withdrawn: withdrawn, Accrual: accrual}
}

// Check Luna namber
func (o *Order) IsValid() (isValid bool) {
	err := goluhn.Validate(o.OrderNr)
	return err == nil

}

func (o *Order) getAccrual() *float64 {
	if o.Accrual.IsZero() {
		return nil
	}
	acc := o.Accrual.InexactFloat64()
	return &acc
}

func (o *Order) getUploded() string {

	return o.Uploaded.Format(time.RFC3339)
}

func (o *Order) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Number  string   `json:"number"`
		Status  string   `json:"status"`
		Accrual *float64 `json:"accrual,omitempty"`
		Uploded string   `json:"uploaded_at"`
	}{
		Number:  o.OrderNr,
		Status:  o.Status.String(),
		Accrual: o.getAccrual(),
		Uploded: o.getUploded(),
	})
}
