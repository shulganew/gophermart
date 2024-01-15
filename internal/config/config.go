package config

import (
	"errors"
	"flag"
	"os"

	"github.com/shulganew/gophermart/internal/api/validators"
	"go.uber.org/zap"
)

type Config struct {
	//flag -a, Market address
	Address string

	// Server loyality address
	Accrual string

	//dsn connection string
	DSN string
}

func InitConfig() *Config {

	config := Config{}
	//read command line argue
	marketAddress := flag.String("a", "localhost:8090", "Service Gophermart address")
	loyaltyAddress := flag.String("r", "localhost:8080", "Service Loyality address")
	dsnf := flag.String("d", "", "Data Source Name for DataBase connection")
	flag.Parse()

	//check and parse URL

	startaddr, startport := validators.CheckURL(*marketAddress)
	accaddr, accport := validators.CheckURL(*loyaltyAddress)

	//save config
	config.Address = startaddr + ":" + startport
	config.Accrual = accaddr + ":" + accport

	//read OS ENVs
	addr, exist := os.LookupEnv(("RUN_ADDRESS"))

	//if env var does not exist  - set def value
	if exist {
		config.Address = addr
		zap.S().Infoln("Set result address from evn RUN_ADDRESS: ", config.Address)

	} else {
		zap.S().Infoln("Env var RUN_ADDRESS not found, use default", config.Address)
	}

	dsn, exist := os.LookupEnv(("DATABASE_URI"))

	//init shotrage DB from env

	if exist {
		zap.S().Infoln("Use DataBase DSN from evn DATABASE_URI, use: ", dsn)
		config.DSN = dsn

	} else if *dsnf != "" {
		dsn = *dsnf
		zap.S().Infoln("Use DataBase from -d flag, use: ", dsn)
		config.DSN = dsn
	} else {
		zap.S().Errorf("Can't make config for DB, set -d flag or DATABASE_URI env for DSN!")
		panic(errors.New("Can't make config for DB, set -d flag or DATABASE_URI env for DSN!"))
	}

	acc, exist := os.LookupEnv(("ACCRUAL_SYSTEM_ADDRESS"))

	if exist {
		config.Accrual = acc
		zap.S().Infoln("Set accrual service addres from evn ACCRUAL_SYSTEM_ADDRESS: ", config.Accrual)

	} else {
		zap.S().Infoln("Env var ACCRUAL_SYSTEM_ADDRESS not found, use default", config.Accrual)
	}

	zap.S().Infoln("Configuration complite")
	return &config
}
