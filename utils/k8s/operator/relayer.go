package operator

import (
	"fmt"

	v1 "github.com/allinbits/starport-operator/api/v1"
)

type RelayerConfig struct {
	NodesetName   string `json:"nodeset_name"`
	FaucetName    string `json:"faucet_name"`
	AccountPrefix string `json:"account_prefix"`
	HDPath        string `json:"hd_path"`
}

func (rc RelayerConfig) Validate() error {
	if rc.NodesetName == "" {
		return fmt.Errorf("missing nodeset name")
	}

	if rc.AccountPrefix == "" {
		return fmt.Errorf("missing account prefix")
	}

	return nil
}
func BuildRelayer(c RelayerConfig) (v1.RelayerChain, error) {
	if err := c.Validate(); err != nil {
		return v1.RelayerChain{}, err
	}

	rcs := v1.RelayerChain{
		Nodeset:       &c.NodesetName,
		AccountPrefix: &c.AccountPrefix,
		HDPath:        &c.HDPath,
	}

	if c.FaucetName != "" {
		rcs.Faucet = &c.FaucetName
	}

	return rcs, nil
}
