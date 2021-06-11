package tx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/allinbits/demeris-backend/api/router/deps"
	// typestx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gin-gonic/gin"
	// "google.golang.org/protobuf/proto"
	// "google.golang.org/protobuf/types/known/anypb"
)

func Register(router *gin.Engine) {
	router.POST("/tx/:chain", Tx)
	router.GET("/tx/ticket/:ticket", GetTicket)
}

// Tx relays a transaction to an internal node for the specified chain.
// @Summary Relays a transaction to the relevant chain.
// @Tags Tx
// @ID tx
// @Description Relays a transaction to the relevant chain.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} TxResponse
// @Failure 500,403 {object} deps.Error
// @Router /tx/{chainName} [post]
func Tx(c *gin.Context) {
	// var tx typestx.Tx
	var tx TxData
	var meta TxMeta

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	err := c.BindJSON(&tx)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("failed to parse JSON"), http.StatusBadRequest)

		d.WriteError(c, e,
			"Failed to parse JSON",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	meta.Chain, err = d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("chain %s does not exist", chainName), http.StatusBadRequest)

		d.WriteError(c, e,
			"Invalid chain",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	err = validateTx(&tx, &meta, d)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("invalid transaction"), http.StatusBadRequest)

		d.WriteError(c, e,
			"invalid transaction",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	txhash, err := relayTx(d, tx, meta)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("relaying tx failed"), http.StatusBadRequest)

		d.WriteError(c, e,
			"relaying tx failed",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, TxResponse{
		Ticket: txhash,
	})
}

// validateTx populates metadata and
func validateTx(tx *TxData, meta *TxMeta, d *deps.Deps) error {

	err := validateSignatures(tx, meta, d)

	if err != nil {
		return err
	}

	err = validateBody(tx, meta, d)

	if err != nil {
		return err
	}

	// TODO: Fetch sequence for ticketing system
	err = validateAuthInfo(tx, meta, d)

	if err != nil {
		return err
	}

	return nil
}

// validateBody validates the data inside the body and populates the relevant metadata
func validateBody(tx *TxData, meta *TxMeta, d *deps.Deps) error {
	for _, m := range tx.TxBody.Messages {
		if m.Type == "/ibc.applications.transfer.v1.MsgTransfer" {
			sourcePort := m.SourcePort
			sourceChannel := m.SourceChannel

			tokenDenom := m.Token.Denom

			if sourcePort != "transfer" {
				return fmt.Errorf("Invalid IBC Port %s", sourcePort)
			}

			if "ibc/" == tokenDenom[:4] {
				tokenHash := tokenDenom[4:]
				denomTrace, err := d.Database.DenomTrace(meta.Chain.ChainName, tokenHash)
				if err != nil {
					return fmt.Errorf("Invalid denom trace")
				}

				channels := strings.Split(denomTrace.Path, "/transfer")
				if channels[0] != sourceChannel {
					return fmt.Errorf("IBC forward is disabled for multi-hop tokens. Try sending it back through the original channel.")
				}
			}
		}
	}

	return nil
}

// validateAuthInfo validates the data inside auth_info and populates the relevant metadata
func validateAuthInfo(tx *TxData, meta *TxMeta, d *deps.Deps) error {

	if infos := tx.AuthInfo.SignerInfos; len(infos) == 1 {
		// Fetch signer sequence
		meta.SignerSequence = tx.AuthInfo.SignerInfos[0].Sequence
	} else {
		return fmt.Errorf("Invalid number of signatures. Expected 1, got %d", len(infos))
	}

	return nil
}

// validateSignatures ensures the signature exists
func validateSignatures(tx *TxData, meta *TxMeta, d *deps.Deps) error {
	if len(tx.Signatures) != 1 {
		return fmt.Errorf("Invalid number of signatures")
	}

	return nil
}

// RelayTx relays the tx to the specifc endpoint
// RelayTx will also perform the ticketing mechanism
// Always expect broadcast mode to be `async`
func relayTx(d *deps.Deps, tx TxData, meta TxMeta) (string, error) {

	var hash string

	data, err := json.Marshal(tx)

	if err != nil {
		return hash, err
	}

	b := bytes.NewReader(data)

	resp, err := http.Post(meta.Chain.NodeInfo.Endpoint, "application/json", b)

	if err != nil {
		return hash, err
	}

	decoder := json.NewDecoder(resp.Body)
	var resdata map[string]interface{}

	err = decoder.Decode(&resdata)

	if err != nil {
		return hash, err
	}

	val, ok := resdata["txhash"]

	hash = val.(string)

	if !ok {
		return hash, fmt.Errorf("failed to read txhash in response")
	}

	err = d.Store.CreateTicket(meta.Chain.ChainName, hash)

	if err != nil {
		return hash, err
	}

	return hash, nil
}

// GetTicket returns the transaction status n.
// @Summary Gets ticket by id.
// @Tags Chain
// @ID chain
// @Description Gets transaction status by ticket id.
// @Param ticketId path string true "ticket id"
// @Produce json
// @Success 200 {object} TxStatus
// @Failure 500,403 {object} deps.Error
// @Router /tx/ticket/{ticketId} [get]
func GetTicket(c *gin.Context) {
	var res TxStatus

	d := deps.GetDeps(c)

	ticketId := c.Param("ticketId")

	ticket, err := d.Store.Get(ticketId)

	if err != nil {
		e := deps.NewError(
			"tx",
			fmt.Errorf("cannot retrieve ticket with id %v", ticketId),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve ticket",
			"id",
			e.ID,
			"name",
			ticketId,
			"error",
			err,
		)

		return
	}

	json.Unmarshal([]byte(ticket), &res)

	c.JSON(http.StatusOK, res)
}
