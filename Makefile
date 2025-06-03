ARGS = "postgres://shortener:shortener@localhost:5434/shortener?sslmode=disable"
PPROFDIR = "./profiles/"
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

.PHONY: build
build:
	go build -ldflags "\
		-X 'main.buildVersion=$(VERSION)' \
		-X 'main.buildDate=$(DATE)' \
		-X 'main.buildCommit=$(COMMIT)'" \
		-o cmd/shortener/shortener cmd/shortener/main.go
	chmod +x cmd/shortener/shortener

fmt:
	@echo "Running gofmt..."
	gofmt -s -w .
impr:
	goimports -local github.com/TimBerk/go-link-shortener -v -w .
check: fmt imports lint test bench

doc:
	godoc -http=:8082 -v
swag:
	swag init -g internal/app/handler/handler.go -o swagger --parseDependency --parseInternal
lint:
	go mod verify
	go vet ./...
	staticcheck ./...
stlint:
	go build -o staticlint ./cmd/staticlint
	staticlint ./...

run:
	go run cmd/shortener/main.go -d $(ARGS)
test:
	go test -count=1 -cover ./...
bench:
	go test -bench=. ./... | grep "^Benchmark"
rtest:
	go test -race ./...
pgr:
	docker-compose up -d
pgs:
	docker-compose stop -d
cpup:
	(PPROF_TMPDIR=${PPROFDIR} go tool pprof -http :8081 http://127.0.0.1:8080/debug/pprof/profile)
mpup:
	(PPROF_TMPDIR=${PPROFDIR} go tool pprof -http :8081 http://127.0.0.1:8080/debug/pprof/heap)


.PHONY: generate
generate:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/shortener.proto
