package rest

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/cns/chainwatch"
	"github.com/allinbits/demeris-backend/utils/validation"
	v1 "github.com/allinbits/starport-operator/api/v1"
	v12 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/allinbits/demeris-backend/utils/k8s"

	"github.com/allinbits/demeris-backend/models"
	"github.com/allinbits/demeris-backend/utils/k8s/operator"
	"github.com/gin-gonic/gin"
)

const addChainRoute = "/add"

type addChainRequest struct {
	models.Chain

	SkipChannelCreation  bool                           `json:"skip_channel_creation"`
	NodeConfig           *operator.NodeConfiguration    `json:"node_config"`
	RelayerConfiguration *operator.RelayerConfiguration `json:"relayer_configuration"`
}

func (r *router) addChainHandler(ctx *gin.Context) {
	newChain := addChainRequest{}

	if err := ctx.ShouldBindJSON(&newChain); err != nil {
		e(ctx, http.StatusBadRequest, validation.MissingFieldsErr(err, false))
		r.s.l.Error("cannot bind input data to Chain struct", err)
		return
	}

	if err := validateFees(newChain.Chain); err != nil {
		e(ctx, http.StatusBadRequest, err)
		r.s.l.Error("fee validation failed", err)
		return
	}

	if err := validateDenoms(newChain.Chain); err != nil {
		e(ctx, http.StatusBadRequest, err)
		r.s.l.Error("fee validation failed", err)
		return
	}

	k := k8s.Querier{
		Client:    *r.s.k,
		Namespace: r.s.defaultK8SNamespace,
	}

	rsrc := schema.GroupResource{Group: "apps.starport.cloud", Resource: "nodesets"}
	testErr := k8serrors.NewNotFound(rsrc, newChain.ChainName)
	r.s.l.Infow("test error this", "test", testErr)

	r.s.l.Infow("debug", "info", r.s.nodesetInformer, "ns", r.s.defaultK8SNamespace, "name", newChain.ChainName)
	nodeset, err := k8s.GetChain(r.s.nodesetInformer, r.s.defaultK8SNamespace, newChain.ChainName)
	r.s.l.Infow("this is nodeset", "nodes", nodeset, "error", err)
	if errors.Is(err, k8serrors.NewNotFound(rsrc, newChain.ChainName)) {
		r.s.l.Infow("trying to add a kubernetes nodeset which is already there, ignoring", "error", err)
		newChain.NodeConfig = nil
	}

	if newChain.NodeConfig != nil {
		newChain.NodeConfig.Namespace = r.s.defaultK8SNamespace

		newChain.NodeConfig.Name = newChain.ChainName

		// we trust that TestnetConfig holds the real chain ID
		if newChain.NodeConfig.TestnetConfig != nil &&
			*newChain.NodeConfig.TestnetConfig.ChainId != newChain.NodeInfo.ChainID {
			newChain.NodeInfo.ChainID = *newChain.NodeConfig.TestnetConfig.ChainId
		}

		newChain.NodeConfig.TracelistenerDebug = r.s.debug

		node, err := operator.NewNode(*newChain.NodeConfig)
		if err != nil {
			e(ctx, http.StatusBadRequest, err)
			r.s.l.Error("cannot add chain", err)
			return
		}

		switch newChain.NodeConfig.DisableMinFeeConfig {
		case true:
			node.Spec.Config.Nodes.TraceStoreContainer.ImagePullPolicy = v12.PullNever
		default:
			minGasPriceVal := newChain.RelayerToken().GasPriceLevels.Low / 2
			minGasPricesStr := fmt.Sprintf("%v%s", minGasPriceVal, newChain.RelayerToken().Name)

			cfgOverride := v1.ConfigOverride{
				App: []v1.TomlConfigField{
					{
						Key: "minimum-gas-prices",
						Value: v1.TomlConfigFieldValue{
							String: &minGasPricesStr,
						},
					},
				},
			}
			node.Spec.Config.Nodes.ConfigOverride = &cfgOverride
		}

		hasFaucet := false
		if node.Spec.Init != nil {
			hasFaucet = node.Spec.Init.Faucet != nil
		}

		if newChain.RelayerConfiguration == nil {
			newChain.RelayerConfiguration = &operator.DefaultRelayerConfiguration
		}

		if err := newChain.RelayerConfiguration.Validate(); err != nil {
			e(ctx, http.StatusBadRequest, err)
			r.s.l.Errorw("cannot validate relayer configuration", "error", err)
			return
		}

		if err := r.s.rc.AddChain(chainwatch.Chain{
			Name:                 newChain.ChainName,
			AddressPrefix:        newChain.NodeInfo.Bech32Config.MainPrefix,
			HasFaucet:            hasFaucet,
			SkipChannelCreation:  newChain.SkipChannelCreation,
			HDPath:               newChain.DerivationPath,
			RelayerConfiguration: *newChain.RelayerConfiguration,
		}); err != nil {
			e(ctx, http.StatusInternalServerError, err)
			r.s.l.Error("cannot add chain name to cache", err)
			return
		}

		r.s.l.Debugw("node config", "config", node)

		if err := k.AddNode(*node); err != nil {
			e(ctx, http.StatusInternalServerError, err)
			r.s.l.Error("cannot add chain", err)
			return
		}
	}

	if err := r.s.d.AddChain(newChain.Chain); err != nil {
		e(ctx, http.StatusInternalServerError, err)
		r.s.l.Error("cannot add chain", err)
		return
	}

	return
}
func (r *router) addChain() (string, gin.HandlerFunc) {
	return addChainRoute, r.addChainHandler
}

func validateFees(c models.Chain) error {
	ft := c.FeeTokens()
	if len(ft) == 0 {
		return fmt.Errorf("no fee token specified")
	}

	for _, denom := range ft {
		if denom.GasPriceLevels.Empty() {
			return fmt.Errorf("fee levels for %s are not defined", denom.Name)
		}
	}

	return nil
}

func validateDenoms(c models.Chain) error {
	foundRelayerDenom := false
	for _, d := range c.Denoms {
		if d.RelayerDenom {
			if foundRelayerDenom {
				return fmt.Errorf("multiple relayer denoms detected")
			}

			if d.MinimumThreshRelayerBalance == nil {
				return fmt.Errorf("relayer denom detected but no relayer minimum threshold balance defined")
			}

			foundRelayerDenom = true
		}
	}

	return nil
}
