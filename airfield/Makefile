GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

cli-create:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/create-airline cmd/create-airline/main.go

cli-lookup:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/lookup cmd/lookup/main.go

