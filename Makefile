OBJS = $(shell find cmd -mindepth 1 -type d -execdir printf '%s\n' {} +)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
BASEPKG = github.com/allinbits/demeris-backend
EXTRAFLAGS :=

.PHONY: $(OBJS) clean generate-swagger

all: $(OBJS)

clean:
	@rm -rf build docs/swagger.* docs/docs.go

generate-swagger:
	go generate ${BASEPKG}/docs
	@rm docs/docs.go

dev-int-tests:
	ENV=dev go test -v ./tests/...

staging-int-tests:
	./test_data/client/staging/port-forward.sh
	ENV=staging go test -v ./tests/...
	pkill -f "port-forward"

$(OBJS):
	go build -o build/$@ -ldflags='-X main.Version=${BRANCH}-${COMMIT}' ${EXTRAFLAGS} ${BASEPKG}/cmd/$@
