module github.com/allinbits/demeris-backend

go 1.16

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/jmoiron/sqlx => github.com/abraithwaite/sqlx v1.3.2-0.20210331022513-df9bf9884350
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/cockroachdb/cockroach-go/v2 v2.1.1
	github.com/containerd/fifo v1.0.0
	github.com/cosmos/cosmos-sdk v0.42.4
	github.com/cosmos/gaia/v4 v4.2.1
	github.com/gin-contrib/zap v0.0.1
	github.com/gin-gonic/gin v1.7.1
	github.com/go-playground/validator/v10 v10.5.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/iamolegga/enviper v1.2.1
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgproto3/v2 v2.0.7 // indirect
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/jackc/pgx/v4 v4.11.0
	github.com/jmoiron/sqlx v1.3.3
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/sony/sonyflake v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/swaggo/swag v1.7.0
	github.com/tendermint/liquidity v1.2.4
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210415154028-4f45737414dc // indirect
	golang.org/x/sys v0.0.0-20210415045647-66c3f260301c // indirect
	golang.org/x/text v0.3.6 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)
