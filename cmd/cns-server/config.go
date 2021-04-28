package main

import (
	"fmt"

	"github.com/allinbits/demeris-backend/utils/configuration"
	"github.com/allinbits/demeris-backend/utils/validation"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	DatabaseConnectionURL string `validate:"required"`
	LogPath               string
	RESTAddress           string `validate:"required"`
	Debug                 bool
}

func (c Config) Validate() error {
	err := validator.New().Struct(c)
	if err == nil {
		return nil
	}
	return fmt.Errorf(
		"configuration file error: %w",
		validation.MissingFieldsErr(err, false),
	)
}

func readConfig() (*Config, error) {
	var c Config

	return &c, configuration.ReadConfig(&c, "demeris-cns", map[string]string{
		"RESTAddress": ":9999",
	})
}
