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
	chainInfo := utils.LoadSignleChainInfo(env, chainName, t)

	var info cns.Chain
	err := json.Unmarshal(chainInfo.Payload, &info)
	if err != nil {
		fmt.Printf("Error while unamrshelling chain info : %v", err)
	}

	addressPrefix := info.NodeInfo.Bech32Config.PrefixAccount

	c, err = CreateChainClient(cc.KeyringServiceName, cc.RPC, addressPrefix, t.TempDir())
	require.NoError(t, err)
	require.NotNil(t, c)

	mnemonic := cc.Mnemonic
	if env == StagingEnvKey {
		mnemonic = GetMnemonic(chainName)

	}

	c.AddressPrefix = addressPrefix
	c.HDPath = info.DerivationPath
	c.Enabled = info.Enabled
	c.ChainName = info.ChainName
	c.Mnemonic = mnemonic

	a, err := c.ImportMnemonic(cc.Key, c.Mnemonic, c.HDPath)
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
