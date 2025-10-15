install:
    go install ./cmd/piper

gen:
    go tool devtool gen --all ./cmd/piper

update:
    go get -u ./...

dep:
    go mod tidy

test:
    go test -failfast ./...

test-race:
    CGO_ENABLED=1 go test -race ./...

fmt-go:
    go tool gofumpt -w -l .
