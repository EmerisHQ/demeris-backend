package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/utils/k8s"

	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc"

	"github.com/gin-gonic/gin"
)

const (
	denomsDataRoute = "/denoms/:chain"
	grpcPort        = 9090
)

type denomsDataResponse struct {
	Denoms        []string `json:"denoms"`
	StakingDenoms []string `json:"staking_denoms"`
}

func (r *router) denomsDataHandler(ctx *gin.Context) {
	chainName := ctx.Param("chain")

	ready, err := k8s.ChainRunning(r.s.nodesetInformer, r.s.defaultK8SNamespace, chainName)
	if err != nil || !ready {
		e(ctx, http.StatusInternalServerError, fmt.Errorf("chain %s not ready", chainName))
		r.s.l.Error("chain not ready", "error", err, "ready value", ready)
		return
	}

	resp, err := queryDenomData(chainName)
	if err != nil {
		e(ctx, http.StatusInternalServerError, err)
		r.s.l.Error("cannot query chain denom data", err)
		return
	}

	ctx.JSON(http.StatusOK, resp)

	return
}

func queryDenomData(chainName string) (denomsDataResponse, error) {
	grpcConn, err := grpc.Dial(fmt.Sprintf("%s:%d", chainName, grpcPort), grpc.WithInsecure())
	if err != nil {
		return denomsDataResponse{}, err
	}

	bankQuery := bank.NewQueryClient(grpcConn)
	stakingQuery := staking.NewQueryClient(grpcConn)

	suppResp, err := bankQuery.TotalSupply(context.Background(), &bank.QueryTotalSupplyRequest{})
	if err != nil {
		return denomsDataResponse{}, fmt.Errorf("cannot query total supply, %w", err)
	}

	stakingDenom, err := stakingQuery.Params(context.Background(), &staking.QueryParamsRequest{})
	if err != nil {
		return denomsDataResponse{}, fmt.Errorf("cannot query staking params, %w", err)
	}

	resp := denomsDataResponse{}

	for _, c := range suppResp.Supply {
		resp.Denoms = append(resp.Denoms, c.Denom)
	}

	resp.StakingDenoms = append(resp.StakingDenoms, stakingDenom.Params.BondDenom)

	return resp, nil
}

func (r *router) denomsData() (string, gin.HandlerFunc) {
	return denomsDataRoute, r.denomsDataHandler
}
