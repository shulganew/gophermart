package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
)

// Get data from Accrual system
func FetchOrderStatus(orderNr string, conf *config.Config) (*model.AccrualResponce, error) {

	client := &http.Client{}

	url, err := url.JoinPath(conf.Accrual, "api", "orders", orderNr)
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
		return nil, errors.New("no correct answer from accural system")
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
