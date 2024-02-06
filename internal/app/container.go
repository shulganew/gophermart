package app

import (
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/ports/client"
	"github.com/shulganew/gophermart/internal/ports/storage"
	"github.com/shulganew/gophermart/internal/services"
)

// A container pattern.
type Container struct {
	calcSrv  *services.CalculationService
	client   *client.Accrual
	userSrv  *services.UserService
	accSrv   *services.AccrualService
	orderSrv *services.OrderService
	conf     *config.Config
}

func NewContainer(conf *config.Config, stor *storage.Repo) *Container {
	container := &Container{}
	container.conf = conf
	container.calcSrv = services.NewCalcService(stor)
	container.userSrv = services.NewUserService(stor)
	container.client = client.NewAccrualClient(conf)
	container.accSrv = services.NewAccrualService(stor, container.client)
	container.orderSrv = services.NewOrderService(stor)
	return container
}

func (c *Container) GetCalculationService() *services.CalculationService {
	return c.calcSrv
}

func (c *Container) GetAccrual() *client.Accrual {
	return c.client
}

func (c *Container) GetUserService() *services.UserService {
	return c.userSrv
}

func (c *Container) GetAccrualService() *services.AccrualService {
	return c.accSrv
}

func (c *Container) GetOrderService() *services.OrderService {
	return c.orderSrv
}

func (c *Container) GetConfig() *config.Config {
	return c.conf
}
