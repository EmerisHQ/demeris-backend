package client

import (
	"encoding/json"
	"os"

	"github.com/allinbits/demeris-backend-models/cns"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	StagingEnvKey      = "staging"
	AkashMnemonicKey   = "AKASH_MNEMONIC"
	CosmosMnemonicKey  = "COSMOS_MNEMONIC"
	TerraMnemonicKey   = "TERRA_MNEMONIC"
	OsmosisMnemonicKey = "OSMOSIS_MNEMONIC"
)

// GetClient is to create client and imports mnemonic and returns created chain client
func GetClient(env string, chainName string, cc ChainClient, dir string) (c *ChainClient, err error) {
	chainInfo, err := utils.LoadSingleChainInfo(env, chainName)
	if err != nil {
		return nil, err
	}

	var info cns.Chain
	err = json.Unmarshal(chainInfo.Payload, &info)
	if err != nil {
		return nil, err
	}

	c, err = CreateChainClient(cc.RPC, cc.KeyringServiceName, info.NodeInfo.ChainID, dir)
	if err != nil {
		return nil, err
	}

	mnemonic := cc.Mnemonic
	if env == StagingEnvKey {
		mnemonic = GetMnemonic(chainName)
	}

	c.AddressPrefix = info.NodeInfo.Bech32Config.PrefixAccount
	c.HDPath = info.DerivationPath
	c.Enabled = info.Enabled
	c.ChainName = info.ChainName
	c.Mnemonic = mnemonic
	c.ChainName = chainName
	if len(info.Denoms) != 0 {
		c.Denom = info.Denoms[0].Name
	}

	return c, nil
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
