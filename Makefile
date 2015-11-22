.PHONY: test

build:
	go build ./cmd/... ./...

test:
	go test -v ./...

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
	go install github.com/mcls/nomad
