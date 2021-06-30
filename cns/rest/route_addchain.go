package rest

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/cns/chainwatch"

	"github.com/allinbits/demeris-backend/utils/validation"

	"github.com/allinbits/demeris-backend/utils/k8s"

	"github.com/allinbits/demeris-backend/models"
	"github.com/allinbits/demeris-backend/utils/k8s/operator"
	"github.com/gin-gonic/gin"
)

const addChainRoute = "/add"

type addChainRequest struct {
	models.Chain

	NodeConfig *operator.NodeConfiguration `json:"node_config"`
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


	k := k8s.Querier{
		Client:    *r.s.k,
		Namespace: r.s.defaultK8SNamespace,
	}

	if _, err := k.ChainByName(newChain.ChainName); !errors.Is(err, k8s.ErrNotFound) {
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

		node, err := operator.NewNode(*newChain.NodeConfig)
		if err != nil {
			e(ctx, http.StatusBadRequest, err)
			r.s.l.Error("cannot add chain", err)
			return
		}

		if err := r.s.rc.AddChain(chainwatch.Chain{
			Name:          newChain.ChainName,
			AddressPrefix: newChain.NodeInfo.Bech32Config.MainPrefix,
			HasFaucet:     node.Spec.Init.Faucet != nil,
			HDPath:        newChain.DerivationPath,
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
	for _, denom := range c.FeeTokens() {
		if denom.FeeLevels.Empty() {
			return fmt.Errorf("fee levels for %s are not defined", denom.Name)
		}
	}

	return nil
}
