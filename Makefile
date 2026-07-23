GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

vuln:
	govulncheck -show verbose ./...


ITERATOR_URI=git:///tmp?exclude=properties.edtf:deprecated=.*
ITERATOR_SOURCE=https://github.com/sfomuseum-data/sfomuseum-data-maps.git

compile:
	@make compile-airfield
	@make compile-architecture
	@make compile-curatorial
	@make compile-maps
	@make compile-geo
	@make compile-wof
	@make compile-placetypes
	@make compile-spatial

compile-spatial:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/sfom-spatial-update-hierarchies cmd/sfom-spatial-update-hierarchies/main.go

compile-placetypes:
	@echo "compile plactypes tools"

compile-wof:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)"  -o bin/sfom-wof-import-features cmd/sfom-wof-import-features/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)"  -o bin/sfom-wof-refresh-features cmd/sfom-wof-refresh-features/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)"  -o bin/sfom-wof-ensure-properties cmd/sfom-wof-ensure-properties/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)"  -o bin/sfom-wof-merge-properties cmd/sfom-wof-merge-properties/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)"  -o bin/sfom-wof-ensure-features cmd/sfom-wof-ensure-features/main.go

compile-geo:
	@make compile-geotag
	@make compile-georef

compile-geotag:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/sfom-geotag-add cmd/sfom-geotag-add/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/sfom-geotag-remove cmd/sfom-geotag-remove/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/sfom-geotag-build-update cmd/sfom-geotag-build-update/main.go

compile-georef:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/sfom-georef-add cmd/sfom-georef-add/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/sfom-georef-remove cmd/sfom-georef-remove/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/sfom-georef-recompile-subject cmd/sfom-georef-recompile-subject/main.go

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


# Lambda

lambda:
	@make lambda-wof

lambda-wof:
	@make lambda-wof-import

lambda-wof-import:
	if test -f bootstrap; then rm -f bootstrap; fi
	if test -f sfom-wof-import-features.zip; then rm -f sfom-wof-import-features.zip; fi
	GOARCH=arm64 GOOS=linux go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -tags lambda.norpc -o bootstrap cmd/sfom-wof-import-features/main.go
	zip sfom-wof-import-features.zip bootstrap
	rm -f bootstrap

# Placetypes

placetypes-spec:
	@echo "please fix me"
	# cp placetypes/placetypes.json dist/placetypes.json.last
	# go run cmd/sfom-placetypes-compile-spec/main.go > dist/placetypes.json.tmp
	# cp dist/placetypes.json.tmp dist/placetypes.json
	# go run cmd/sfom-placetypes-render-spec/main.go -path placetypes/docs/images/placetypes.png

# Maps

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

# subject (object):
# https://collection.sfomuseum.org/objects/1897902471/
# https://static.sfomuseum.org/data/189/790/247/1/1897902471.geojson
#
# depiction (image):
# https://collection.sfomuseum.org/images/1897903961/
# https://github.com/sfomuseum-data/sfomuseum-data-media-collection/blob/main/data/189/790/396/1/1897903961.geojson
#
# reference (bangkok):
# https://spelunker.whosonfirst.org/id/102025263

debug-georef-photo:
	go run -mod $(GOMOD) cmd/assign-georeferences/main.go \
		-depiction-reader-uri repo://$(CWD)/fixtures/sfomuseum-data-media-collection \
		-depiction-writer-uri stdout:// \
		-subject-reader-uri repo://$(CWD)/fixtures/sfomuseum-data-collection \
		-subject-writer-uri stdout:// \
		-depiction-id 1897903961 \
		-reference sfomuseum:depicts=102025263


# subject (object):
# https://collection.sfomuseum.org/objects/1511907389
# https://static.sfomuseum.org/data/151/190/738/9/1511907389.geojson
#
# depiction (image):
# https://collection.sfomuseum.org/images/1527829811/
# https://static.sfomuseum.org/data/152/782/981/1/1527829811.geojson
#
# reference (noumea)
# https://spelunker.whosonfirst.org/id/890413117
# reference (sydney)
# https://spelunker.whosonfirst.org/id/101932003

debug-georef-flightcover:
	mkdir -p $(CWD)/fixtures/debug/data
	rm -rf $(CWD)/fixtures/debug/data/*
	go run -mod $(GOMOD) cmd/assign-georeferences/main.go \
		-depiction-reader-uri repo://$(CWD)/fixtures/sfomuseum-data-media-collection \
		-depiction-writer-uri repo://$(CWD)/fixtures/debug \
		-subject-reader-uri repo://$(CWD)/fixtures/sfomuseum-data-collection \
		-subject-writer-uri repo://$(CWD)/fixtures/debug \
		-depiction-id 1527829811 \
		-reference sfomuseum:flightcover_to=890413117 \
		-reference sfomuseum:flightcover_from=101932003 \
		-verbose
