package services

// User registration service
type Market struct {
	stor MarketPlaceholder
}

type MarketPlaceholder interface {
	GetOrder()
}

func NewMarket(stor MarketPlaceholder) *Market {
	return &Market{}
}

func (r *Market) GetOrder() {

}
