PIPER = go run ./cmd/piper --log-level=debug

build:
	$(PIPER) do go build

