TEST_COMMAND = go test

.PHONY: build
build:
	go build -v

.PHONY: test
test: 
	$(TEST_COMMAND) -v -cover -parallel 5 -failfast  ./... 

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	revive -formatter friendly -config revive.toml ./...

.PHONY: spell
spell:
	misspell -error ./**

.PHONY: staticcheck
staticcheck:
	staticcheck ./...

.PHONY: gosec
gosec:
	gosec -tests ./...

.PHONY: inspect
inspect: lint spell gosec staticcheck
