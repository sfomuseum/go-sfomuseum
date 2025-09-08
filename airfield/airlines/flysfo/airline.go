package flysfo

import (
	"fmt"
)

type Airline struct {
	WhosOnFirstId int64  `json:"wof:id"`
	Name          string `json:"wof:name"`
	FlysfoID      string `json:"flysfo:airline_id"`
	IATACode      string `json:"iata:code,omitempty"`
	ICAOCode      string `json:"icao:code,omitempty"`
}

func (a *Airline) String() string {
	return fmt.Sprintf("%s %s \"%s\" %d", a.IATACode, a.ICAOCode, a.Name, a.WhosOnFirstId)
}
