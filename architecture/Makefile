GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w


cli-lookup:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/lookup cmd/lookup/main.go

cli-complex:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" --tags json1 -o bin/current-complex cmd/current-complex/main.go

