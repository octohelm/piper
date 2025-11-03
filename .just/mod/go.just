devtool := 'go tool devtool'

install:
    go install ./cmd/piper

fmt:
    {{ devtool }} fmt -w -l .

gen:
    {{ devtool }} gen --all ./cmd/piper

update:
    go get -u ./...

dep:
    go mod tidy

test:
    go test -count=1 -failfast ./...

test-race:
    CGO_ENABLED=1 \
      go test -count=1  -race ./...
