package test_utils

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const settingsPath = "./test_data/%s/settings.yaml"

type clientConfig struct {
	Timeout    int16 `yaml:"timeout"`
	TcpTimeout int16 `yaml:"tcp_timeout"`
	SslTimeout int16 `yaml:"ssl_timeout"`
}

func CreateNetClient(env string, t *testing.T) *http.Client {

	require.NotEmpty(t, env)

	yFile, err := ioutil.ReadFile(fmt.Sprintf(settingsPath, env))
	require.NoError(t, err)

	conf := clientConfig{}
	err = yaml.Unmarshal(yFile, &conf)
	require.NoError(t, err)

	netTransport := &http.Transport{
		// FIXME: Replace Dial with DialContext
		Dial: (&net.Dialer{
			Timeout: time.Duration(conf.TcpTimeout) * time.Millisecond,
		}).Dial,
		TLSHandshakeTimeout: time.Duration(conf.SslTimeout) * time.Millisecond,
	}

	netClient := &http.Client{
		Timeout:   time.Millisecond * time.Duration(conf.Timeout),
		Transport: netTransport,
	}

	return netClient
}
