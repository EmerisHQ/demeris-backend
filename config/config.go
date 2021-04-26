package config

import (
	"fmt"
	"net/url"

	"github.com/allinbits/navigator-utils/validation"

	"github.com/allinbits/navigator-utils/configuration"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	DatabaseConnectionURL string `validate:"required"`
	ListenAddr            string `validate:"required"`
	CNSAddr               string `validate:"required"`
	Debug                 bool
}

func (c Config) Validate() error {
	err := validator.New().Struct(c)
	if err != nil {
		return validation.MissingFieldsErr(err, false)
	}

	u, err := url.ParseRequestURI(c.CNSAddr)
	if err != nil {
		return fmt.Errorf("invalid url, %w", err)
	}

	switch {
	case u.Scheme == "":
		return fmt.Errorf("missing scheme")
	case u.Host == "":
		return fmt.Errorf("missing hostname")
	}

	return nil
}

func Read() (*Config, error) {
	var c Config

	return &c, configuration.ReadConfig(&c, "navigator-api", map[string]string{
		"ListenAddr": ":9090",
	})
}
