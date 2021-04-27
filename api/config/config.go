package config

import (
	"github.com/allinbits/demeris-backend/utils/validation"

	"github.com/allinbits/demeris-backend/utils/configuration"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	DatabaseConnectionURL string `validate:"required"`
	ListenAddr            string `validate:"required"`
	CNSAddr               string `validate:"required,url"`
	Debug                 bool
}

func (c Config) Validate() error {
	err := validator.New().Struct(c)
	if err != nil {
		return validation.MissingFieldsErr(err, false)
	}

	return nil
}

func Read() (*Config, error) {
	var c Config

	return &c, configuration.ReadConfig(&c, "navigator-api", map[string]string{
		"ListenAddr": ":9090",
	})
}
