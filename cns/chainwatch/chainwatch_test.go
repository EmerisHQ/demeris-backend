package chainwatch

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.uber.org/zap"

	"github.com/allinbits/demeris-backend/models"
	v1 "github.com/allinbits/starport-operator/api/v1"
)

func TestInstance_updatePrimaryChannelForChain(t *testing.T) {
	tests := []struct {
		name      string
		chainsMap map[string]models.Chain
		relayer   v1.Relayer
		want      map[string]models.Chain
	}{
		{
			"no previous primary channels",
			map[string]models.Chain{
				"cosmos-hub-testnet": {
					ChainName: "cosmos-hub",
				},
				"akash-testnet": {
					ChainName: "akash",
				},
			},
			v1.Relayer{
				Status: v1.RelayerStatus{
					Paths: []v1.RelayerPath{
						{
							"cosmos-hub-testnet": v1.RelayerSide{
								ChannelID: "ch0",
							},
							"akash-testnet": v1.RelayerSide{
								ChannelID: "ch0",
							},
						},
					},
				},
			},
			map[string]models.Chain{
				"cosmos-hub-testnet": {
					ChainName: "cosmos-hub",
					PrimaryChannel: models.DbStringMap{
						"akash": "ch0",
					},
				},
				"akash-testnet": {
					ChainName: "akash",
					PrimaryChannel: models.DbStringMap{
						"cosmos-hub": "ch0",
					},
				},
			},
		},
		{
			"previous primary channels were defined among two chains",
			map[string]models.Chain{
				"cosmos-hub-testnet": {
					ChainName: "cosmos-hub",
					PrimaryChannel: models.DbStringMap{
						"akash": "ch0",
					},
				},
				"akash-testnet": {
					ChainName: "akash",
					PrimaryChannel: models.DbStringMap{
						"cosmos-hub": "ch0",
					},
				},
			},
			v1.Relayer{
				Status: v1.RelayerStatus{
					Paths: []v1.RelayerPath{
						{
							"cosmos-hub-testnet": v1.RelayerSide{
								ChannelID: "ch1",
							},
							"akash-testnet": v1.RelayerSide{
								ChannelID: "ch1",
							},
						},
					},
				},
			},
			map[string]models.Chain{
				"cosmos-hub-testnet": {
					ChainName: "cosmos-hub",
					PrimaryChannel: models.DbStringMap{
						"akash": "ch0",
					},
				},
				"akash-testnet": {
					ChainName: "akash",
					PrimaryChannel: models.DbStringMap{
						"cosmos-hub": "ch0",
					},
				},
			},
		},
		{
			"two chains already connected, one isn't",
			map[string]models.Chain{
				"cosmos-hub-testnet": {
					ChainName: "cosmos-hub",
					PrimaryChannel: models.DbStringMap{
						"akash": "ch0",
					},
				},
				"akash-testnet": {
					ChainName: "akash",
					PrimaryChannel: models.DbStringMap{
						"cosmos-hub": "ch0",
					},
				},
				"persistence-testnet": {
					ChainName: "persistence",
				},
			},
			v1.Relayer{
				Status: v1.RelayerStatus{
					Paths: []v1.RelayerPath{
						{
							"cosmos-hub-testnet": v1.RelayerSide{
								ChannelID: "ch0",
							},
							"akash-testnet": v1.RelayerSide{
								ChannelID: "ch0",
							},
						},
						{
							"cosmos-hub-testnet": v1.RelayerSide{
								ChannelID: "ch2",
							},
							"persistence-testnet": v1.RelayerSide{
								ChannelID: "ch0",
							},
						},
						{
							"persistence-testnet": v1.RelayerSide{
								ChannelID: "ch1",
							},
							"akash-testnet": v1.RelayerSide{
								ChannelID: "ch2",
							},
						},
					},
				},
			},
			map[string]models.Chain{
				"cosmos-hub-testnet": {
					ChainName: "cosmos-hub",
					PrimaryChannel: models.DbStringMap{
						"akash":       "ch0",
						"persistence": "ch2",
					},
				},
				"akash-testnet": {
					ChainName: "akash",
					PrimaryChannel: models.DbStringMap{
						"cosmos-hub":  "ch0",
						"persistence": "ch2",
					},
				},
				"persistence-testnet": {
					ChainName: "persistence",
					PrimaryChannel: models.DbStringMap{
						"cosmos-hub": "ch0",
						"akash":      "ch1",
					},
				},
			},
		},
		{
			"three chains already connected, paths report different channels, no variations should happen on already present channels",
			map[string]models.Chain{
				"cosmos-hub-testnet": {
					ChainName: "cosmos-hub",
					PrimaryChannel: models.DbStringMap{
						"akash":       "ch0",
						"persistence": "ch2",
					},
				},
				"akash-testnet": {
					ChainName: "akash",
					PrimaryChannel: models.DbStringMap{
						"cosmos-hub":  "ch0",
						"persistence": "ch2",
					},
				},
				"persistence-testnet": {
					ChainName: "persistence",
					PrimaryChannel: models.DbStringMap{
						"cosmos-hub": "ch0",
						"akash":      "ch1",
					},
				},
			},
			v1.Relayer{
				Status: v1.RelayerStatus{
					Paths: []v1.RelayerPath{
						{
							"cosmos-hub-testnet": v1.RelayerSide{
								ChannelID: "ch1",
							},
							"akash-testnet": v1.RelayerSide{
								ChannelID: "ch1",
							},
						},
						{
							"cosmos-hub-testnet": v1.RelayerSide{
								ChannelID: "ch2",
							},
							"persistence-testnet": v1.RelayerSide{
								ChannelID: "ch0",
							},
						},
						{
							"persistence-testnet": v1.RelayerSide{
								ChannelID: "ch1",
							},
							"akash-testnet": v1.RelayerSide{
								ChannelID: "ch2",
							},
						},
					},
				},
			},
			map[string]models.Chain{
				"cosmos-hub-testnet": {
					ChainName: "cosmos-hub",
					PrimaryChannel: models.DbStringMap{
						"akash":       "ch0",
						"persistence": "ch2",
					},
				},
				"akash-testnet": {
					ChainName: "akash",
					PrimaryChannel: models.DbStringMap{
						"cosmos-hub":  "ch0",
						"persistence": "ch2",
					},
				},
				"persistence-testnet": {
					ChainName: "persistence",
					PrimaryChannel: models.DbStringMap{
						"cosmos-hub": "ch0",
						"akash":      "ch1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl, _ := zap.NewDevelopment()

			in := Instance{
				l: zl.Sugar(),
			}

			for k, v := range tt.chainsMap {
				if v.PrimaryChannel == nil {
					v.PrimaryChannel = models.DbStringMap{}
					tt.chainsMap[k] = v
				}
			}
			res := in.updatePrimaryChannelForChain(tt.chainsMap, tt.relayer)

			require.Equal(t, tt.want, res)
		})
	}
}
