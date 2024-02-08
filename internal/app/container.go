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

func (c *Application) GetAccrual() *client.Accrual {
	return c.client
}

func (c *Application) GetUserService() *services.UserService {
	return c.userSrv
}

func (c *Application) GetAccrualService() *services.AccrualService {
	return c.accSrv
}

func (c *Application) GetOrderService() *services.OrderService {
	return c.orderSrv
}

func (c *Application) GetConfig() *config.Config {
	return c.conf
}

func (c *Application) GetRepo() *storage.Repo {
	return c.stor
}
