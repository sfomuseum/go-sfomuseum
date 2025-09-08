package sfomuseum

type Flight struct {
	WhosOnFirstId     int64  `json:"wof:id"`
	SFOMuseumFlightId string `json:"sfomuseum:flight_id"`
}
