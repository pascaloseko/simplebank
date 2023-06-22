.PHONE: help
help:
	@printf "%24s: %s\n" "proto-compile" "Compile protobuf schemas"
	@echo "Examples:"
	@printf "%24s: %s\n" "bash" "Opens bash inside Docker (accepts CMD_ARGS)"
	@printf "%24s: %s\n" "clean" "Deletes all temporary files"
	@printf "%24s: %s\n" "fmt" "Runs go fmt"
	@printf "%24s: %s\n" "lint" "Runs all linters"
	@printf "%24s: %s\n" "proto-compile" "Compile protobuf schemas"
	@printf "%24s: %s\n" "tidy" "Run go mod tidy"
	@printf "%24s: %s\n" "test" "Runs the test suite (accepts CMD_ARGS)"
	@echo

.PHONY: lint
lint:
	@docker compose run --rm app golangci-lint run -v

.PHONY: fmt
fmt:
	@docker compose run --rm app gofmt -d -w -s .


.PHONY: bash
bash:
	@docker compose run --rm app bash ${CMD_ARGS}

.PHONY: test
test:
	@docker compose run --rm app ./run-tests ${CMD_ARGS}

.PHONY: clean
clean:
	@docker compose run --rm app go clean -cache

.PHONY: tidy
tidy:
	@docker compose run --rm app go mod tidy

.PHONY: generate
generate:
	@docker compose run --rm app sqlc generate

.PHONY: mock
mock:
	@docker compose run --rm app mockgen --package mockdb --destination repo/mock/store.go github.com/simplebank/repo Store

.PHONY: server
server:
	@docker compose run --rm app go run main.go server