package test_utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"gopkg.in/yaml.v3"
)

const settingsPath = "./test_data/%s/settings.yaml"

type clientConfig struct {
	Timeout    int16 `yaml:"timeout"`
	TcpTimeout int16 `yaml:"tcp_timeout"`
	SslTimeout int16 `yaml:"ssl_timeout"`
}

type HttpClient struct {
	*http.Client
	scheme   string
	host     string
	basePath string
}

func NewHttpClient(env, scheme, host, basePath string) (*HttpClient, error) {
	httpClient, err := createNetClient(env)
	if err != nil {
		return nil, err
	}

	return &HttpClient{
		Client:   httpClient,
		scheme:   scheme,
		host:     host,
		basePath: basePath,
	}, nil
}

func (c *HttpClient) BuildUrl(path string, args ...interface{}) string {
	path = fmt.Sprintf(path, args...)

	u := url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.basePath + path,
	}

	return u.String()
}

// GetJson performs a GET request and unmarshalls jsonOut.
func (c *HttpClient) GetJson(jsonOut interface{}, endpoint string, endpointParams ...interface{}) error {
	return c.DoJson("GET", nil, jsonOut, endpoint, endpointParams...)
}

// DoJson performs a HTTP request sending jsonIn as body and unmarshalling jsonOut.
func (c *HttpClient) DoJson(method string, jsonIn interface{}, jsonOut interface{}, endpoint string, endpointParams ...interface{}) error {
	url := c.BuildUrl(endpoint, endpointParams...)

	var requestBodyReader io.Reader
	if jsonIn != nil {
		requestBodyBytes, err := json.Marshal(jsonIn)
		if err != nil {
			return fmt.Errorf("marshalling json body: %w", err)
		}
		requestBodyReader = bytes.NewReader(requestBodyBytes)
	}

	req, err := http.NewRequest(method, url, requestBodyReader)
	if err != nil {
		return fmt.Errorf("preparing http request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("during http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading http response body: %w", err)
	}

	err = json.Unmarshal(body, jsonOut)
	if err != nil {
		return fmt.Errorf("unmarshalling http response body: %w. Body was: %s", err, body)
	}

	return nil
}

func createNetClient(env string) (*http.Client, error) {

	if env == "" {
		return nil, fmt.Errorf("got nil ENV env")
	}

	yFile, err := ioutil.ReadFile(fmt.Sprintf(settingsPath, env))
	if err != nil {
		return nil, err
	}

	conf := clientConfig{}
	err = yaml.Unmarshal(yFile, &conf)
	if err != nil {
		return nil, err
	}

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

	return netClient, nil
}
