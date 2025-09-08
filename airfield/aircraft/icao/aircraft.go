package icao

import (
	"fmt"
)

type Aircraft struct {
	ModelFullName       string
	Description         string
	WTC                 string
	Designator          string
	ManufacturerCode    string
	AircraftDescription string
	EngineCount         string
	EngineType          string
}

func (a *Aircraft) String() string {
	return fmt.Sprintf("%s %s \"%s\"", a.ManufacturerCode, a.Designator, a.ModelFullName)
}
