package denom

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/allinbits/demeris-backend/models"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/denom/verify-trace/:chain/:hash", VerifyTrace)
}

func VerifyTrace(c *gin.Context) {
	var res VerifiedTraceResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"denom/verify-trace",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chain := c.Param("chain")
	hash := c.Param("hash")

	denomTrace, err := d.Database.DenomTrace(chain, hash)

	if err != nil {
		handleFailedQuery(c, d, err, hash, chain)
	}

	res.IbcDenom = fmt.Sprintf("ibc/%s", hash)
	res.Path = denomTrace.Path

	// check if the path uses only the supported `transfer` port.

	channels := strings.Split(res.Path, "/transfer")

	for idx, channel := range channels {
		ch := strings.Trim(channel, "/")

		// port other than transfer being used
		if strings.Contains(ch, "/") {

			err = errors.New(fmt.Sprintf("Unsupported path %s", res.Path))

			e := deps.NewError(
				"denom/verify-trace",
				fmt.Errorf("invalid denom %v with path %v", hash, res.Path),
				http.StatusBadRequest,
			)

			c.Error(e)

			d.Logger.Errorw(
				"invalid denom",
				"id",
				e.ID,
				"hash",
				hash,
				"path",
				res.Path,
				"err",
				err,
			)
		}

		channels[idx] = ch
	}

	var client models.Client
	var chainInfo models.Chain
	var trace Trace

	for idx, channel := range channels {

		client, _ = d.Database.QueryIBCClientTrace(chain, channel)

		chainInfo, _ = d.Database.Chain(chain)

		if counterparty := chainInfo.CounterpartyNames[client.ChannelId]; counterparty == "" {
			err = errors.New(fmt.Sprintf("Unsupported client id %s on chain %s", client.ChannelId, chain))

			e := deps.NewError(
				"denom/verify-trace",
				fmt.Errorf("Unsupported client id when resolving path for %s", hash),
				http.StatusBadRequest,
			)

			c.Error(e)

			d.Logger.Errorw(
				"invalid client id",
				"id",
				e.ID,
				"hash",
				hash,
				"path",
				res.Path,
				"client_id",
				client.ClientId,
				"chain",
				chain,
				"err",
				err,
			)
		} else {
			trace.ChainName = chain
			trace.Channel = client.ChannelId
			trace.ClientId = client.ClientId
			trace.CounterpartyName = counterparty

			// query counterparty chain name
			client, _ = d.Database.QueryIBCClientTrace(chain, channel)

		}

		res.Trace = append(res.Trace, trace)
	}

	c.JSON(http.StatusOK, gin.H{
		"trace": res,
	})
}

func handleFailedQuery(c *gin.Context, d *deps.Deps, err error, hash string, chain string) {
	e := deps.NewError(
		"denom/verify-trace",
		fmt.Errorf("cannot query token hash %v on chain %v", hash, chain),
		http.StatusBadRequest,
	)

	c.Error(e)

	d.Logger.Errorw(
		"cannot query database for denom",
		"id",
		e.ID,
		"hash",
		hash,
		"chain",
		chain,
		"error",
		err,
	)

	return
}
