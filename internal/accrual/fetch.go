package accrual

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
)

type AccrualClient struct {
	conf *config.Config
}

func NewAccrualClient(conf *config.Config) *AccrualClient {

	return &AccrualClient{conf: conf}
}

// Get data from Accrual system
func (a AccrualClient) FetchOrderStatus(orderNr string) (*model.AccrualResponce, error) {

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
	var accResp model.AccrualResponce
	err = json.NewDecoder(res.Body).Decode(&accResp)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return &accResp, nil
}
