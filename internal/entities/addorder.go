package entities

import "github.com/shopspring/decimal"

type AddOrder struct {
	UserID     string
	OrderNr    string
	IsPreOrder bool
	Withdrawn  decimal.Decimal
}

func NewAddOrder(userID string, orderNr string, isPreOrder bool, withdrawn decimal.Decimal) *AddOrder {
	return &AddOrder{UserID: userID, OrderNr: orderNr, IsPreOrder: isPreOrder, Withdrawn: withdrawn}
}
