package test_utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	chainsFolderPath       = "./ci/%s/chains/"
	jsonSuffix             = ".json"
	enabledKey             = "enabled"
	nameKey                = "chain_name"
	clientChainsFolderPath = "./test_data/client/"
)

type EnvChain struct {
	Name    string
	Enabled bool
	Payload []byte
}

func LoadChainsInfo(env string, t *testing.T) []EnvChain {

	require.NotEmpty(t, env)

	d := fmt.Sprintf(chainsFolderPath, env)
	files, err := ioutil.ReadDir(d)
	require.NoError(t, err)

	var chains []EnvChain
	for _, f := range files {
		if strings.HasSuffix(f.Name(), jsonSuffix) {
			jFile, err := ioutil.ReadFile(d + f.Name())
			require.NoError(t, err)

			temp := map[string]interface{}{}
			err = json.Unmarshal(jFile, &temp)
			require.NoError(t, err)

			ch := EnvChain{}
			ch.Payload = jFile
			ch.Enabled = temp[enabledKey].(bool)
			ch.Name = temp[nameKey].(string)
			chains = append(chains, ch)
		}
	}

	return chains
}

func LoadClientChainsInfo(t *testing.T) []EnvChain {

	d := clientChainsFolderPath
	files, err := ioutil.ReadDir(d)
	require.NoError(t, err)

	var chains []EnvChain
	for _, f := range files {
		if strings.HasSuffix(f.Name(), jsonSuffix) {
			jFile, err := ioutil.ReadFile(d + f.Name())
			require.NoError(t, err)

			temp := map[string]interface{}{}
			err = json.Unmarshal(jFile, &temp)
			require.NoError(t, err)

			ch := EnvChain{}
			ch.Payload = jFile
			ch.Enabled = temp[enabledKey].(bool)
			ch.Name = temp[nameKey].(string)
			chains = append(chains, ch)
		}
	}

	return chains
}
