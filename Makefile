GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

compile:
	@make compile-airfield
	@make compile-architecture
	@make compile-curatorial

compile-airfield:
	@make compile-airlines
	@make compile-airports
	@make compile-aircraft

compile-architecture:
	@make compile-gates
	@make compile-galleries
	@make compile-terminals

compile-curatorial:
	@make compile-publicart-data
	@make compile-exhibitions-data
	@make compile-collection-data

compile-airlines:
	go run -mod $(GOMOD) cmd/compile-flysfo-airlines-data/main.go \
		-iterator-uri 'git:///tmp?exclude=properties.edtf:deprecated=.*' \
		https://github.com/sfomuseum-data/sfomuseum-data-enterprise.git
	go run -mod $(GOMOD) cmd/compile-sfomuseum-airlines-data/main.go  \
		-iterator-uri 'git:///tmp?exclude=properties.edtf:deprecated=.*' \
		https://github.com/sfomuseum-data/sfomuseum-data-enterprise.git

compile-airports:
	go run -mod $(GOMOD) cmd/compile-sfomuseum-airports-data/main.go \
		-verbose \
		-iterator-uri 'git:///tmp?include=properties.sfomuseum:placetype=airport&exclude=properties.edtf:deprecated=.*' \
		https://github.com/sfomuseum-data/sfomuseum-data-whosonfirst.git

compile-aircraft:
	go run -mod $(GOMOD) cmd/compile-sfomuseum-aircraft-data/main.go \
		-iterator-uri 'git:///tmp?exclude=properties.edtf:deprecated=.*' \
		https://github.com/sfomuseum-data/sfomuseum-data-aircraft.git

compile-gates:
	go run -mod $(GOMOD) -ldflags="$(LDFLAGS)" cmd/compile-gates-data/main.go

compile-terminals:
	go run -mod $(GOMOD) -ldflags="$(LDFLAGS)" cmd/compile-terminals-data/main.go

compile-galleries:
	go run -mod $(GOMOD) -ldflags="$(LDFLAGS)" cmd/compile-galleries-data/main.go

compile-publicart-data:
	go run -mod $(GOMOD) -ldflags="-s -w" cmd/compile-publicart-data/main.go

compile-exhibitions-data:
	go run -mod $(GOMOD) -ldflags="-s -w" cmd/compile-exhibitions-data/main.go

compile-collection-data:
	go run -mod $(GOMOD) -ldflags="-s -w" cmd/compile-collection-data/main.go
