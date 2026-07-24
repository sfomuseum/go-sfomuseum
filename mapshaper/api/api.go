package api

import (
	"github.com/sfomuseum/go-sfomuseum/mapshaper"
)

type MapshaperAPIOptions struct {
	Mapshaper      *mapshaper.Mapshaper
	UploadsMaxSize int64
}
