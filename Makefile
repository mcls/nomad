.PHONY: test

test:
	go test -v ../nomad

# go get golang.org/x/tools/cmd/cover
cover:
	go test -covermode=count -coverprofile=`pwd`/coverage.out && go tool cover -html=`pwd`/coverage.out
