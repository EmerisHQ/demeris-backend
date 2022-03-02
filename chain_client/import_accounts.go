package client

import (
	"encoding/json"
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
func GetClient(t *testing.T, env string, chainName string, cc Client) (c *Client) {
	chainInfo, err := utils.LoadSingleChainInfo(env, chainName)
	require.NoError(t, err)

	var info cns.Chain
	err = json.Unmarshal(chainInfo.Payload, &info)
	require.NoError(t, err)

	addressPrefix := info.NodeInfo.Bech32Config.PrefixAccount
	chainID := info.NodeInfo.ChainID

	c, err = CreateChainClient(cc.RPC, cc.KeyringServiceName, chainID, t.TempDir())
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
	c.ChainName = chainName
	if len(info.Denoms) != 0 {
		c.Denom = info.Denoms[0].Name
	}

	a, err := c.ImportMnemonic(cc.Key, c.Mnemonic, c.HDPath)
	require.NoError(t, err)
	require.NotNil(t, a)

	return c
}

// GetMnemonic returns the mnemonic of particular chain for staging accounts
func GetMnemonic(chName string) string {
	var mnemonic string

	switch chName {
	case "akash":
		mnemonic = os.Getenv(AkashMnemonicKey)
	case "cosmos-hub":
		mnemonic = os.Getenv(CosmosMnemonicKey)
	case "terra":
		mnemonic = os.Getenv(TerraMnemonicKey)
	case "osmosis":
		mnemonic = os.Getenv(OsmosisMnemonicKey)
	default:
		mnemonic = os.Getenv("MNEMONIC")
	}

	return mnemonic
}
