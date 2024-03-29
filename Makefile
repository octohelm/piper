PIPER = go run ./cmd/piper

DEBUG = 0
ifeq ($(DEBUG),1)
	PIPER := GRAPH=1 TTY=0 $(PIPER) --log-level=debug
endif

tidy:
	$(PIPER) mod tidy

build:
	$(PIPER) do go build

archive:
	$(PIPER) do go archive

ship.build:
	$(PIPER) do ship build

ship:
	$(PIPER) do ship publish

ship.multi-builder:
	PIPER_BUILDER_HOST="tcp://arm64builder@?platform=linux/arm64,docker-image://amd64builder@?platform=linux/amd64" \
		$(PIPER) do ship

release:
	$(PIPER) do release

gen:
	go run ./internal/cmd/tool gen ./cmd/piper

dep.update:
	go get -u ./...

install:
	go install ./cmd/piper
