package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/shulganew/gophermart/internal/app/config"
	"github.com/shulganew/gophermart/internal/entities"
	"go.uber.org/zap"
)

type Accrual struct {
	conf *config.Config
}

func NewAccrualClient(conf *config.Config) *Accrual {
	return &Accrual{conf: conf}
}

// Get data from Accrual system.
func (a Accrual) GetOrderStatus(orderNr string) (*entities.AccrualResponce, error) {
	client := &http.Client{}

	url, err := url.JoinPath(a.conf.Accrual, "api", "orders", orderNr)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("no correct answer from accural system: " + strconv.Itoa(res.StatusCode))
	}

	//Load data to AccrualResponce from json
	var accResp entities.AccrualResponce
	err = json.NewDecoder(res.Body).Decode(&accResp)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := res.Body.Close()
		zap.S().Errorln("Can't close response body: ", err)
	}()

	return &accResp, nil
}
