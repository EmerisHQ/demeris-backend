package test_utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	chainclient "github.com/allinbits/demeris-backend/chainclient"
	"github.com/emerishq/demeris-backend-models/cns"
)

const (
	chainsFolderPath       = "./ci/%s/chains/"
	jsonSuffix             = ".json"
	clientChainsFolderPath = "./test_data/client/%s/"
)

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
			if err = json.Unmarshal(jFile, &chain); err != nil {
				return nil, err
			}

			chains = append(chains, chain)
		}
	}

	return chains, nil
}

func LoadClientChainsInfo(env string) ([]chainclient.ChainClient, error) {
	if env == "" {
		return nil, fmt.Errorf("got nil ENV env")
	}

	d := fmt.Sprintf(clientChainsFolderPath, env)
	files, err := ioutil.ReadDir(d)
	if err != nil {
		return nil, err
	}

	var chains []chainclient.ChainClient
	for _, f := range files {
		if strings.HasSuffix(f.Name(), jsonSuffix) {
			jFile, err := ioutil.ReadFile(d + f.Name())
			if err != nil {
				return nil, err
			}

			var ch chainclient.ChainClient
			if err = json.Unmarshal(jFile, &ch); err != nil {
				return nil, err
			}

			chains = append(chains, ch)
		}
	}

	return chains, nil
}
