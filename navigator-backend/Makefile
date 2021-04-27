.PHONY: build gen-go

all: build

build:
	 go build -v --ldflags="-s -w"  github.com/allinbits/navigator-backend/cmd/navigator-api

gen-go:
	openapi-generator generate -i swagger.yml -g go-gin-server --git-repo-id navigator-backend --git-user-id allinbits -o oapi