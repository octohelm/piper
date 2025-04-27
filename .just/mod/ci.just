export TTY := "0"
piper := 'go tool piper' + if env("DEBUG", "0") == '1' { ' --log-level=debug' } else { '' }

build:
    {{ piper }} do go build

archive:
    {{ piper }} do go archive

tidy:
    {{ piper }} mod tidy

release:
    {{ piper }} do release

ship:
    {{ piper }} do ship piper push

ship-distroless:
    {{ piper }} do ship distroless push

ship-multi-builder:
    export PIPER_BUILDER_HOST := 'tcp://arm64builder@?platform=linux/arm64,docker-image://amd64builder@?platform=linux/amd64'
    {{ piper }} do ship
