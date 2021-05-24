package rpcwatcher

import (
	"github.com/allinbits/demeris-backend/utils/configuration"
	"github.com/allinbits/demeris-backend/utils/validation"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	DatabaseConnectionURL string `validate:"required"`
	TendermintNode        string `validate:"required"`
	RedisURL              string `validate:"required"`
	Debug                 bool
}

func (c *Config) Validate() error {
	err := validator.New().Struct(c)
	if err == nil {
		return nil
	}

	return validation.MissingFieldsErr(err, false)
}

func ReadConfig() (*Config, error) {
	var c Config
	return &c, configuration.ReadConfig(&c, "rpcwatcher", map[string]string{
		"TendermintNode": "http://localhost:26657",
		"RedisURL":       "http://localhost:6379",
	})
}
