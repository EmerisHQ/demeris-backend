module github.com/allinbits/demeris-backend

go 1.16

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/jmoiron/sqlx => github.com/abraithwaite/sqlx v1.3.2-0.20210331022513-df9bf9884350
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
	k8s.io/client-go => k8s.io/client-go v0.21.1
)

require (
	github.com/allinbits/starport-operator v0.0.1-alpha.18
	github.com/cockroachdb/cockroach-go/v2 v2.1.1
	github.com/containerd/fifo v1.0.0
	github.com/cosmos/cosmos-sdk v0.42.5
	github.com/cosmos/gaia/v4 v4.2.1
	github.com/cssivision/reverseproxy v0.0.1
	github.com/ethereum/go-ethereum v1.10.3
	github.com/gin-contrib/zap v0.0.1
	github.com/gin-gonic/gin v1.7.1
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator/v10 v10.5.0
	github.com/go-redis/redis/v8 v8.8.3
	github.com/golang/snappy v0.0.3 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/iamolegga/enviper v1.2.1
	github.com/jackc/pgproto3/v2 v2.0.7 // indirect
	github.com/jackc/pgx/v4 v4.11.0
	github.com/jmoiron/sqlx v1.3.3
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lib/pq v1.10.1
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/r3labs/diff v1.1.0
	github.com/sony/sonyflake v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/swag v1.7.0
	github.com/tendermint/liquidity v1.2.4
	github.com/tendermint/tendermint v0.34.10
	github.com/zsais/go-gin-prometheus v0.1.0
	go.uber.org/zap v1.17.0
	golang.org/x/crypto v0.0.0-20210415154028-4f45737414dc // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/grpc v1.37.0
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.9.0
)
