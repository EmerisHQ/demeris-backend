package test_utils

import (
	"fmt"
	"io/ioutil"
	"strings"

	netv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

type EmerisIngress struct {
	Host            string
	Protocol        string
	APIServerPath   string
	LiquidityPath   string
	PriceOraclePath string
}

type EmerisAdminIngress struct {
	Host          string
	Protocol      string
	CNSServerPath string
}

const (
	ingressFilePath = "./ci/%s/ingress.yaml"
	emerisValue     = "emeris"
	liquidityValue  = "liquidity"
	hubValue        = "cosmos-hub"
	oracleValue     = "price-oracle"
	cnsValue        = "cns-server"
)

func LoadIngressInfo(env string) (EmerisIngress, EmerisAdminIngress, error) {
	emIngress := EmerisIngress{}
	emAdminIngress := EmerisAdminIngress{}

	if env == "" {
		return emIngress, emAdminIngress, fmt.Errorf("got nil ENV env")
	}

	yFile, err := ioutil.ReadFile(fmt.Sprintf(ingressFilePath, env))
	if err != nil {
		return emIngress, emAdminIngress, err
	}

	// Original sample: https://dx13.co.uk/articles/2021/01/15/kubernetes-types-using-go/
	decoder := scheme.Codecs.UniversalDeserializer()

	for _, resourceYAML := range strings.Split(string(yFile), "---") {

		// skip empty documents, `Decode` will fail on them
		if len(resourceYAML) == 0 {
			continue
		}

		obj, _, err := decoder.Decode(
			[]byte(resourceYAML),
			nil,
			nil)
		if err != nil {
			return emIngress, emAdminIngress, err
		}

		ingress := obj.(*netv1.Ingress)

		switch ingress.Name {
		case emerisValue:
			initEmerisStruct(ingress, &emIngress)
		default:
			initEmerisAdminStruct(ingress, &emAdminIngress)
		}
	}

	return emIngress, emAdminIngress, nil
}

func initEmerisStruct(data *netv1.Ingress, retIngress *EmerisIngress) {

	retIngress.Host = data.Spec.Rules[0].Host
	if len(data.Spec.TLS) > 0 {
		retIngress.Protocol = "https"
	} else {
		retIngress.Protocol = "http"
	}
	for _, path := range data.Spec.Rules[0].HTTP.Paths {
		normalPath := path.Path[:strings.IndexByte(path.Path, '(')] + "/"
		switch path.Backend.Service.Name {
		case hubValue: // different lquidity aliases across envs
		case liquidityValue:
			retIngress.LiquidityPath = normalPath
		case oracleValue:
			retIngress.PriceOraclePath = normalPath
		default:
			retIngress.APIServerPath = normalPath
		}
	}
}

func initEmerisAdminStruct(data *netv1.Ingress, retIngress *EmerisAdminIngress) {

	retIngress.Host = data.Spec.Rules[0].Host
	if len(data.Spec.TLS) > 0 {
		retIngress.Protocol = "https"
	} else {
		retIngress.Protocol = "http"
	}
	for _, path := range data.Spec.Rules[0].HTTP.Paths {
		// ignore ingress paths which are not Emeris-specific (e.g. websocket on staging)
		if !strings.Contains(path.Path, "(") {
			continue
		}
		normalPath := path.Path[:strings.IndexByte(path.Path, '(')] + "/"
		if path.Backend.Service.Name == cnsValue {
			retIngress.CNSServerPath = normalPath
		}
	}
}
