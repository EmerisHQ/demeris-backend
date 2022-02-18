package experiments

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/allinbits/demeris-api-server/api/account"
)

func TestHelloWorld(t *testing.T) {
	assert := assert.New(t)

	balance := account.Balance{
		Address: "sample",
	}

	assert.Equal("sample", balance.Address)
}
