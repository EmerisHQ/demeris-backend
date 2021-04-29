//go:generate go run github.com/swaggo/swag/cmd/swag i -g ../docs/swagger_gen.go -d ../api --parseDependency -o ./

package docs

// We keep this import here to make sure go mod doesn't remove swaggo dependency,
// otherwise we cannot use the generate statement up there ^.
import _ "github.com/swaggo/swag"
