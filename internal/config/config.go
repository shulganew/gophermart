package config

import (
	"flag"
	"os"
	"time"

	"github.com/shulganew/gophermart/internal/api/validators"
	"go.uber.org/zap"
)

// Check Acceral service every X sec.
const CheckAccrual = 1

const DataBaseType = "postgres"

const TokenExp = time.Hour * 3600

type Config struct {
	//flag -a, Market address
	Address string

	// Server loyality address
	Accrual string

	//dsn connection string
	DSN string

	DSNMitration string

	PassJWT string

	// Eneble databese goose migratios
	Migrations bool
}

func InitConfig() *Config {
	config := Config{}
	//read command line argue
	marketAddress := flag.String("a", "localhost:8088", "Service Gophermart address")
	loyaltyAddress := flag.String("r", "localhost:8090", "Service Loyality address")
	dsnf := flag.String("d", "", "Data Source Name for DataBase connection")
	authJWT := flag.String("p", "JWTsecret", "JWT private key")
	migration := flag.String("mp", "postgresql://postgres:postgres@localhost/postgres?sslmode=disable", "Data Source Name for DataBase connection for migrations (postgres admin user)")
	isMigrations := flag.Bool("m", false, "Enable database migration durnig app start")

	flag.Parse()

	// check and parse URL

	startaddr, startport := validators.CheckURL(*marketAddress)
	accaddr, accport := validators.CheckURL(*loyaltyAddress)

	// save config
	config.Address = startaddr + ":" + startport
	config.Accrual = "http://" + accaddr + ":" + accport

	// read OS ENVs
	addr, exist := os.LookupEnv(("RUN_ADDRESS"))

	// JWT password for users auth
	config.PassJWT = *authJWT

	// if env var does not exist  - set def value
	if exist {
		config.Address = addr
		zap.S().Infoln("Set result address from evn RUN_ADDRESS: ", config.Address)
	} else {
		zap.S().Infoln("Env var RUN_ADDRESS not found, use default", config.Address)
	}

	// set config DSN for postgres admin for database creation
	config.DSNMitration = *migration
	config.Migrations = *isMigrations

	dsn, exist := os.LookupEnv(("DATABASE_URI"))

	// init shotrage DB from env

	if exist {
		zap.S().Infoln("Use DataBase DSN from evn DATABASE_URI, use: ", dsn)
		config.DSN = dsn
	} else if *dsnf != "" {
		dsn = *dsnf
		zap.S().Infoln("Use DataBase from -d flag, use: ", dsn)
		config.DSN = dsn
	} else {
		zap.S().Errorf("Can't make config for DB, set -d flag or DATABASE_URI env for DSN!")
		os.Exit(65)
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
