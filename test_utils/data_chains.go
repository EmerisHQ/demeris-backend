package test_utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

const (
	chainsFolderPath = "./ci/%s/chains/"
	jsonSuffix       = ".json"
	enabledKey       = "enabled"
	nameKey          = "chain_name"
)

type EnvChain struct {
	Name    string
	Enabled bool
	Payload []byte
}

func LoadChainsInfo(env string) ([]EnvChain, error) {

	if env == "" {
		return nil, fmt.Errorf("got nil ENV env")
	}

	d := fmt.Sprintf(chainsFolderPath, env)
	files, err := ioutil.ReadDir(d)
	if err != nil {
		return nil, err
	}

	var chains []EnvChain
	for _, f := range files {
		if strings.HasSuffix(f.Name(), jsonSuffix) {
			jFile, err := ioutil.ReadFile(d + f.Name())
			if err != nil {
				return nil, err
			}

			temp := map[string]interface{}{}
			err = json.Unmarshal(jFile, &temp)
			if err != nil {
				return nil, err
			}

			ch := EnvChain{}
			ch.Payload = jFile
			ch.Enabled = temp[enabledKey].(bool)
			ch.Name = temp[nameKey].(string)
			chains = append(chains, ch)
		}
	}

	return chains, nil
}
