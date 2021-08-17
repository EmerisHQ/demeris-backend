package config

import (
	"time"

	"github.com/allinbits/demeris-backend/utils/configuration"
	"github.com/allinbits/demeris-backend/utils/validation"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	DatabaseConnectionURL string `validate:"required"`
	ListenAddr            string `validate:"required"`
	Debug                 bool
	LogPath               string
	Interval              string   `validate:"required"`
	Whitelistfiats        []string `validate:"required"`
	//Not currently used, but may be used in the future
	CoinmarketcapapiKey string `validate:"required"`
	Fixerapikey         string `validate:"required"`
}

func (c Config) Validate() error {
	err := validator.New().Struct(c)
	if err != nil {
		return validation.MissingFieldsErr(err, false)
	}
	_, err = time.ParseDuration(c.Interval)
	if err != nil {
		return err
	}

	return nil
}

func Read() (*Config, error) {
	var c Config

	return &c, configuration.ReadConfig(&c, "demeris-price-oracle", map[string]string{})
}
