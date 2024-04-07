package main

// DiveRecord is a set of dive parameters denoted in human-readable format. It can be constructed from human input
// or unmarshalled from structured data (JSON). Data may be read from a network request (e.g. HTTP form / JSON object),
// or from the disk (e.g. JSON file / relational database). It can be marshalled to a structured format (JSON).
// When loaded into memory, it holds the data referenced by application's model of a dive (see `Dive`).
type DiveRecord struct {
	DateTime string   `json:"date_time"` // mandatory fields
	Duration Duration `json:"duration"`  //
	Site     string   `json:"site"`      //

	AirTemp           float32 `json:"air_temp"`            // optional fields
	Altitude          uint    `json:"altitude"`            //
	AvgDepth          float32 `json:"avg_depth"`           //
	BodyOfWater       string  `json:"body_of_water"`       //
	CNSEnd            uint    `json:"cns_end"`             //
	CNSStart          uint    `json:"cns_start"`           //
	Current           string  `json:"current"`             //
	DecoAlgFactor     string  `json:"deco_alg_factor"`     //
	DecoDive          bool    `json:"deco_dive"`           //
	DiveComputer      string  `json:"dive_computer"`       //
	Entry             string  `json:"entry"`               //
	Gas               string  `json:"gas"`                 //
	Geo               string  `json:"geo"`                 //
	MaxDepth          float32 `json:"max_depth"`           //
	NightDive         bool    `json:"night_dive"`          //
	Note              string  `json:"note"`                //
	O2                uint    `json:"o2"`                  //
	Operator          string  `json:"operator"`            //
	PerfectWeight     bool    `json:"perfect_weight"`      //
	Suit              string  `json:"suit"`                //
	TankPressureEnd   uint    `json:"tank_pressure_end"`   //
	TankPressureStart uint    `json:"tank_pressure_start"` //
	TankType          string  `json:"tank_type"`           //
	Visibility        string  `json:"visibility"`          //
	WaterMaxTemp      float32 `json:"water_max_temp"`      //
	WaterMinTemp      float32 `json:"water_min_temp"`      //
	Weather           string  `json:"weather"`             //
	Weights           uint    `json:"weights"`             //
}
