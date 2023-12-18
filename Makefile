PIPER = go run ./cmd/piper

DEBUG = 0
ifeq ($(DEBUG),1)
	PIPER := $(PIPER) --log-level=debug
endif

build:
	$(PIPER) do go build

archive:
	$(PIPER) do go archive

release:
	$(PIPER) do release

gen:
	go run ./internal/cmd/tool gen ./cmd/piper

install:
	go install ./cmd/piper
