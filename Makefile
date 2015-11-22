.PHONY: test

default: test build

build:
	@go build ./cmd/...
	@go build ./...

test: setup_test
	go test -v ./...

setup_test:
	rm -r dummy_migrations/

autotest:
	fswatch -o --exclude dummy_migrations ./ | xargs -n1 -I{} make test

# go get golang.org/x/tools/cmd/cover
cover:
	go test -covermode=count -coverprofile=`pwd`/coverage.out && go tool cover -html=`pwd`/coverage.out

# All dependencies
alldeps:
	go list -f '{{join .Deps "\n"}}'

vet:
	go tool vet -v ./

install:
	go install ./cmd/... ./...
