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

staging-int-tests: telepresence
	$(TELEPRESENCE) connect \
		--kubeconfig $(KUBECONFIG)
		--namespace emeris \
		-- go test -v ./tests/... \

# find or download telepresence
.PHONY: telepresence
telepresence: TELEPRESENCE_VERSION?=2.3.7
telepresence:
ifeq (, $(wildcard $(CURRENT_DIR)/bin/telepresence))
	@{ \
	set -e ;\
	echo "Installing telepresence to $(CURRENT_DIR)/bin" ;\
	mkdir -p $(CURRENT_DIR)/bin ;\
	curl -sfL https://app.getambassador.io/download/tel2/$(UNAME)/amd64/$(TELEPRESENCE_VERSION)/telepresence -o $(CURRENT_DIR)/bin/telepresence ;\
	chmod a+x $(CURRENT_DIR)/bin/telepresence ;\
	}
endif
TELEPRESENCE=$(CURRENT_DIR)/bin/telepresence

$(OBJS):
	go build -o build/$@ -ldflags='-X main.Version=${BRANCH}-${COMMIT}' ${EXTRAFLAGS} ${BASEPKG}/cmd/$@
