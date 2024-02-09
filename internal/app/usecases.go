package app

import (
	"github.com/shulganew/gophermart/internal/app/config"
	"github.com/shulganew/gophermart/internal/ports/client"
	"github.com/shulganew/gophermart/internal/ports/storage"
	"github.com/shulganew/gophermart/internal/services"
)

// A container pattern.
type UseCases struct {
	stor     *storage.Repo
	conf     *config.Config
	calcSrv  *services.CalculationService
	client   *client.Accrual
	userSrv  *services.UserService
	accSrv   *services.AccrualService
	orderSrv *services.OrderService
}

func NewUseCases(conf *config.Config, stor *storage.Repo) *UseCases {
	cases := &UseCases{}
	cases.conf = conf
	cases.calcSrv = services.NewCalcService(stor)
	cases.userSrv = services.NewUserService(stor)
	cases.client = client.NewAccrualClient(conf)
	cases.accSrv = services.NewAccrualService(stor, cases.client)
	cases.orderSrv = services.NewOrderService(stor)
	cases.stor = stor

	return cases
}

func (c *UseCases) CalculationService() *services.CalculationService {
	return c.calcSrv
}

func (c *UseCases) Accrual() *client.Accrual {
	return c.client
}

func (c *UseCases) UserService() *services.UserService {
	return c.userSrv
}

func (c *UseCases) AccrualService() *services.AccrualService {
	return c.accSrv
}

func (c *UseCases) OrderService() *services.OrderService {
	return c.orderSrv
}

func (c *UseCases) Config() *config.Config {
	return c.conf
}

func (c *UseCases) Repo() *storage.Repo {
	return c.stor
}
