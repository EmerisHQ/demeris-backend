package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseConnectionURL string `validate:"required"`
	LogPath               string `validate:"required"`
	RESTAddress           string `validate:"required"`
	Debug                 bool
}

func (c Config) Validate() error {
	err := validator.New().Struct(c)
	if err == nil {
		return nil
	}
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return err
	}

	missingFields := []string{}
	for _, e := range ve {
		switch e.Tag() {
		case "required":
			missingFields = append(missingFields, e.StructField())
		}
	}

	return fmt.Errorf("missing configuration file fields: %v", strings.Join(missingFields, ", "))
}

func readConfig() (*Config, error) {
	viper.SetDefault("LogPath", "./navigator-cns.log")
	viper.SetDefault("RESTAddress", ":9999")

	viper.SetConfigName("navigator-cns")
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/navigator-cns/")
	viper.AddConfigPath("$HOME/.navigator-cns")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("config error: %s \n", err)
	}

	return &c, c.Validate()
}
