GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

vuln:
	govulncheck -show verbose ./...


ITERATOR_URI=git:///tmp?exclude=properties.edtf:deprecated=.*
ITERATOR_SOURCE=https://github.com/sfomuseum-data/sfomuseum-data-maps.git

maps-refresh:
	@make compile-maps
	@make maps_catalog
	@make maps_tile_connections

maps-refresh-local:
	@make compile-maps
	@make maps-catalog ITERATOR_URI=repo://?exclude=properties.edtf:deprecated=.* ITERATOR_SOURCE=/usr/local/data/sfomuseum-data-maps
	@make maps-tile-connections  ITERATOR_URI=repo://?exclude=properties.edtf:deprecated=.* ITERATOR_SOURCE=/usr/local/data/sfomuseum-data-maps

maps-catalog:	
	./bin/sfom-maps-catalog-js -iterator-uri $(ITERATOR_URI) -iterator-source $(ITERATOR_SOURCE) > dist/sfomuseum.maps.catalog.js

maps-tile-connections:
	./bin/qgis-tile-connections -iterator-uri $(ITERATOR_URI) -iterator-source $(ITERATOR_SOURCE) > dist/sfomuseum.maps.tileconnections.xml


compile:
	@make compile-airfield
	@make compile-architecture
	@make compile-curatorial
	@make compile-maps

compile-maps:
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/sfom-maps-catalog-js cmd/sfom-maps-catalog-js/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/sfom-maps-qgis-tile-connections cmd/sfom-maps-qgis-tile-connections/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/sfom-maps-new cmd/sfom-maps-new/main.go

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
