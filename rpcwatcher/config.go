package rpcwatcher

import (
	"github.com/allinbits/demeris-backend/utils/configuration"
	"github.com/allinbits/demeris-backend/utils/validation"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	DatabaseConnectionURL string `validate:"required"`
	RedisURL              string `validate:"required,url"`
	ApiURL                string `validate:"required,url"`
	Debug                 bool
	JSONLogs              bool
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
		"RedisURL": "redis-master:6379",
		"ApiURL":   "http://api-server:8000",
	})
}
