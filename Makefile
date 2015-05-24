.PHONY: test

test:
	go test -v ../nomad

cover:
	go test -coverprofile=`pwd`/coverage.out && go tool cover -html=`pwd`/coverage.out
