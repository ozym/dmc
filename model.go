package dmc

import (
	"encoding/json"
	"net"
	"time"
)

type ModelType int

const (
	UnknownModel ModelType = iota
	RadioModel
	CellularModel
	RouterModel
	DataloggerModel
	StrongModel
	GNSSModel
)

type State struct {
	Values map[string]interface{}
}

//{
/*
	Model string `json:"model"`

		Name     *string `json:"name,omitempty"`
		Location *string `json:"location,omitempty"`

		Model   *string `json:"model,omitempty"`
		Version *string `json:"version,omitempty"`
		Serial  *string `json:"serial,omitempty"`
		License *string `json:"license,omitempty"`

		Firmware *string `json:"firmware,omitempty"`
		Software *string `json:"software,omitempty"`
		Macaddr  *string `json:"macaddr,omitempty"`
*/
//}

func (s *State) Marshal() []byte {

	r, err := json.MarshalIndent(s.Values, "", "  ")
	if err != nil {
		panic(err)
	}

	return r
}

func (s *State) String() string {
	r, err := json.Marshal(s.Values)
	if err != nil {
		panic(err)
	}

	return (string)(r)
}

/*
type SOH map[string]interface{}
*/

/*
type SOH struct {
	Status *string `json:"status,omitempty"`

	Supply      *float64 `json:"supply,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`

	Frequency  *float64 `json:"frequency,omitempty"`
	Signal     *float64 `json:"signal,omitempty"`
	Noise      *float64 `json:"noise,omitempty"`
	Reflective *float64 `json:"reflective,omitempty"`
	Ratio      *float64 `json:"ratio,omitempty"`

	RX      *uint32 `json:"rx,omitempty"`
	TX      *uint32 `json:"tx,omitempty"`
	Dropped *uint32 `json:"dropped,omitempty"`
	Bad     *uint32 `json:"bad,omitempty"`
}
*/

/*
func (s *SOH) Marshal() string {

	r, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		panic(err)
	}

	return (string)(r)
}
*/

// The Model inteface ....
type Model interface {
	// Provide a short model name.
	Name() string
	// Allow matching of model names.
	MatchString(name string) bool

	// The Model is part of one or more groups.
	Groups() []ModelType
	// Is the Model a member of a ModelType group.
	Group(group ModelType) bool

	Identify(orig string, ip net.IP, timeout time.Duration, retries int) (*State, error)
}

// ModelList allows looping over the set of defined device Models.
var ModelList = []Model{
	&MikroTik{},
	&Ubiquiti{},
	&Freewave{},
	&ViPR{},
	&Hongdian{},
	&Quanterra{},
	&Cusp{},
	&Rock{},
	&Trimble{},
}
