package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/allinbits/demeris-backend-models/cns"
)

const (
	chainsFolderPath = "./ci/%s/chains/"
	jsonSuffix       = ".json"
)

func LoadSingleChainInfo(env string, chainName string) (cns.Chain, error) {
	d := fmt.Sprintf(chainsFolderPath, env)
	fileName := fmt.Sprintf("%s%s", chainName, jsonSuffix)

	var chain cns.Chain
	jFile, err := ioutil.ReadFile(d + fileName)
	if err != nil {
		return chain, err
	}

	if err := json.Unmarshal(jFile, &chain); err != nil {
		return chain, err
	}

	return chain, nil
}
