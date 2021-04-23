API_BRANCH=$(shell ./branch.sh ../findy-agent-api/)

drop_api:
	go mod edit -dropreplace github.com/findy-network/findy-agent-api

drop_all: drop_api

repl_api:
	go mod edit -replace github.com/findy-network/findy-agent-api=../findy-agent-api

repl_all: repl_api

modules:
	@echo Syncing modules for work brances ...
	go get github.com/findy-network/findy-agent-api@$(API_BRANCH)

deps:
	go get -t ./...

build:
	go build ./...

vet:
	go vet ./...

shadow:
	@echo Running govet
	go vet -vettool=$(GOPATH)/bin/shadow ./...
	@echo Govet success

check_fmt:
	$(eval GOFILES = $(shell find . -name '*.go'))
	@gofmt -l $(GOFILES)

lint:
	@golangci-lint run

lint_e:
	@$(GOPATH)/bin/golint ./... | grep -v export | cat

test:
	go test -v -p 1 -failfast ./...

logged_test:
	go test -v -p 1 -failfast ./... -args -logtostderr=true -v=10

test_cov:
	go test -v -p 1 -failfast -coverprofile=c.out ./... && go tool cover -html=c.out

check: check_fmt vet shadow

dbuild:
	docker build \
		--build-arg HTTPS_PREFIX=$(HTTPS_PREFIX) \
		-t findy-agent-auth \
		.
