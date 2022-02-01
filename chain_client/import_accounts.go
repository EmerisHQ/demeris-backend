package client

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/allinbits/demeris-backend-models/cns"
	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/stretchr/testify/require"
)

const (
	StagingEnvKey      = "staging"
	AkashMnemonicKey   = "AKASH_MNEMONIC"
	CosmosMnemonicKey  = "COSMOS_MNEMONIC"
	TerraMnemonicKey   = "TERRA_MNEMONIC"
	OsmosisMnemonicKey = "OSMOSIS_MNEMONIC"
)

// GetClient is to create client and imports mnemonic and returns created chain client
func GetClient(t *testing.T, env string, chainName string, cc Client) (c Client) {
	var mnemonic string
	chainInfo := utils.LoadSignleChainInfo(env, chainName, t)

	var info cns.Chain
	err := json.Unmarshal(chainInfo.Payload, &info)
	if err != nil {
		fmt.Printf("Error while unamrshelling chain info : %v", err)
	}

	cc.AddressPrefix = info.NodeInfo.Bech32Config.PrefixAccount
	cc.HDPath = info.DerivationPath

	if env == StagingEnvKey {
		mnemonic = GetMnemonic(chainName)
		if mnemonic != "" {
			cc.Mnemonic = mnemonic
		}
	}

	c, err = CreateChainClient(cc.KeyringServiceName, cc.RPC, cc.AddressPrefix, t.TempDir())
	require.NoError(t, err)
	require.NotNil(t, c)

	a, err := c.ImportMnemonic(cc.Key, cc.Mnemonic, cc.HDPath)
	require.NoError(t, err)
	require.NotNil(t, a)

	return c
}

// GetMnemonic returns the mnemonic of particular chain for staging accounts
func GetMnemonic(chName string) string {
	var mnemonic string

	if chName == "akash" {
		mnemonic = os.Getenv(AkashMnemonicKey)
	} else if chName == "cosmos-hub" {
		mnemonic = os.Getenv(CosmosMnemonicKey)
	} else if chName == "terra" {
		mnemonic = os.Getenv(TerraMnemonicKey)
	} else if chName == "osmosis" {
		mnemonic = os.Getenv(OsmosisMnemonicKey)
	} else {
		mnemonic = os.Getenv("MNEMONIC")
	}

	return mnemonic
}
