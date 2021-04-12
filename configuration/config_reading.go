package configuration

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Validator is an object that implements a validation method, which accepts no argument and returns an error.
type Validator interface {
	Validate() error
}

// ReadConfig reads the TOML configuration file in predefined standard paths into v, returns an error if v.Validate()
// returns error, or some configuration file reading error happens.
// v is the destination struct, configName is the name used for the configuration file.
func ReadConfig(v Validator, configName string, defaultValues map[string]string) error {
	for k, v := range defaultValues {
		viper.SetDefault(k, v)
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath(fmt.Sprintf("/etc/%s", configName))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", configName))
	viper.AddConfigPath(".")
	viper.SetEnvPrefix(strings.ToLower(configName))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&v); err != nil {
		return fmt.Errorf("config error: %s \n", err)
	}

	return v.Validate()
}
