package app

import (
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/ports/client"
	"github.com/shulganew/gophermart/internal/ports/storage"
	"github.com/shulganew/gophermart/internal/services"
)

// A container pattern.
type Application struct {
	stor     *storage.Repo
	calcSrv  *services.CalculationService
	client   *client.Accrual
	userSrv  *services.UserService
	accSrv   *services.AccrualService
	orderSrv *services.OrderService
	conf     *config.Config
}

func NewApp(conf *config.Config, stor *storage.Repo) *Application {
	application := &Application{}
	application.conf = conf
	application.calcSrv = services.NewCalcService(stor)
	application.userSrv = services.NewUserService(stor)
	application.client = client.NewAccrualClient(conf)
	application.accSrv = services.NewAccrualService(stor, application.client)
	application.orderSrv = services.NewOrderService(stor)
	application.stor = stor

	return application
}

func (c *Application) CalculationService() *services.CalculationService {
	return c.calcSrv
}

func (c *Application) Accrual() *client.Accrual {
	return c.client
}

func (c *Application) UserService() *services.UserService {
	return c.userSrv
}

func (c *Application) AccrualService() *services.AccrualService {
	return c.accSrv
}

func (c *Application) OrderService() *services.OrderService {
	return c.orderSrv
}

func (c *Application) Config() *config.Config {
	return c.conf
}

func (c *Application) Repo() *storage.Repo {
	return c.stor
}
