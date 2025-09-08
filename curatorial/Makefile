GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

cli:
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/lookup cmd/lookup/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/assign-exhibition-gallery cmd/assign-exhibition-gallery/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/supersede-exhibition cmd/supersede-exhibition/main.go

