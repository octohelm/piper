PIPER = TTY=0 go run ./cmd/piper

DEBUG = 0
ifeq ($(DEBUG),1)
	PIPER := $(PIPER) --log-level=debug
endif

export PIPER_BUILDER_HOST =

tidy:
	$(PIPER) mod tidy

build:
	$(PIPER) do go build

archive:
	$(PIPER) do go archive

ship.build:
	$(PIPER) do ship build

ship:
	$(PIPER) do ship piper push

ship.distroless:
	$(PIPER) do ship distroless push

ship.multi-builder:
	PIPER_BUILDER_HOST="tcp://arm64builder@?platform=linux/arm64,docker-image://amd64builder@?platform=linux/amd64" \
		$(PIPER) do ship

fmt:
	gofumpt -w -l .

release:
	$(PIPER) do release

gen:
	go run ./internal/cmd/tool gen ./cmd/piper

dep.update:
	go get -u ./...

install:
	go install ./cmd/piper

test:
	go test -v -failfast ./...

debug.dagger:
	TTY=0 go run ./pkg/dagger/__debug__

debug.distroless:
	$(PIPER) do ship distroless export linux/arm64

debug.kubepkg:
	$(PIPER) do pkg

debug.kubepkg.ocitar:
	$(PIPER) do pkg ocitar