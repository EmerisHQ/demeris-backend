gen-go:
	openapi-generator generate -i swagger.yml -g go-gin-server --git-repo-id navigator-backend --git-user-id allinbits -o oapi