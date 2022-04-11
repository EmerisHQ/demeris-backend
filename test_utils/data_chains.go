package test_utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/allinbits/demeris-backend-models/cns"
	chainClient "github.com/allinbits/demeris-backend/chain_client"
)

const (
	chainsFolderPath       = "./ci/%s/chains/"
	jsonSuffix             = ".json"
	enabledKey             = "enabled"
	nameKey                = "chain_name"
	clientChainsFolderPath = "./test_data/client/%s/"
)

type EnvChain struct {
	Name    string
	Enabled bool
	Payload []byte
}

func LoadChainsInfo(env string) ([]cns.Chain, error) {
	if env == "" {
		return nil, fmt.Errorf("got nil ENV env")
	}

	d := fmt.Sprintf(chainsFolderPath, env)
	files, err := ioutil.ReadDir(d)
	if err != nil {
		return nil, err
	}

	var chains []cns.Chain
	for _, f := range files {
		if strings.HasSuffix(f.Name(), jsonSuffix) {
			jFile, err := ioutil.ReadFile(d + f.Name())
			if err != nil {
				return nil, err
			}

			var chain cns.Chain
			err = json.Unmarshal(jFile, &chain)
			if err != nil {
				return nil, err
			}
			chains = append(chains, chain)
		}
	}

	return chains, nil
}

func LoadClientChainsInfo(env string) ([]chainClient.ChainClient, error) {
	if env == "" {
		return nil, fmt.Errorf("got nil ENV env")
	}

	d := fmt.Sprintf(clientChainsFolderPath, env)
	files, err := ioutil.ReadDir(d)
	if err != nil {
		return nil, err
	}

	var chains []chainClient.ChainClient
	for _, f := range files {
		if strings.HasSuffix(f.Name(), jsonSuffix) {
			jFile, err := ioutil.ReadFile(d + f.Name())
			if err != nil {
				return nil, err
			}

			var ch chainClient.ChainClient
			err = json.Unmarshal(jFile, &ch)
			if err != nil {
				return nil, err
			}
			chains = append(chains, ch)
		}
	}

	return chains, nil
}
