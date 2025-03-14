export PIPER_BUILDER_HOST := ''
export TTY := "0"
piper := 'go tool piper' + if env("DEBUG", "0") == '1' { ' --log-level=debug' } else { '' }
gofumpt := 'go tool gofumpt'

fmt:
    {{ gofumpt }} -w -l .

tidy:
    {{ piper }} mod tidy

build:
    {{ piper }} do go build

archive:
    {{ piper }} do go archive

release:
    {{ piper }} do release

gen:
    go tool devtool gen --all ./cmd/piper

dep:
    go get ./...

update:
    go get -u ./...

install:
    go install ./cmd/piper

test:
    go test -v -failfast ./...

ship:
    {{ piper }} do ship piper push

ship-distroless:
    {{ piper }} do ship distroless push

ship-multi-builder:
    export PIPER_BUILDER_HOST := 'tcp://arm64builder@?platform=linux/arm64,docker-image://amd64builder@?platform=linux/amd64'
    {{ piper }} do ship
