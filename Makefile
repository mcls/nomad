.PHONY: test

GO=$(shell command -v go)
FSWATCH=$(shell command -v fswatch)

default: test build

build:
	@$(GO) build ./cmd/...
	@$(GO) build ./...

test: setup_test
	$(GO) test ./...

setup_test:
	-rm -r dummy_migrations/

autotest:
	$(FSWATCH) -o --exclude dummy_migrations ./ | xargs -n1 -I{} $(MAKE) test

# go get golang.org/x/tools/cmd/cover
cover:
	$(GO) test -covermode=count -coverprofile=`pwd`/coverage.out && $(GO) tool cover -html=`pwd`/coverage.out

# All dependencies
alldeps:
	$(GO) list -f '{{join .Deps "\n"}}'

vet:
	$(GO) tool vet -v ./

install:
	$(GO) install ./cmd/... ./...
