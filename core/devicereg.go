package core

import "github.com/eensymachines-in/scheduling"

// DevReg : represents one device registration in the database
type DevReg struct {
	Serial    string                       `json:"serial" bson:"serial"` // unique serial number of the device
	Schedules []*scheduling.JSONRelayState `json:"schedules" bson:"schedules"`
}

// DevRegHttpPayload : this one is the device registration as a simple http payload incoming
// When registering a new device all that we need is a serial number and the list of relay ids
type DevRegHttpPayload struct {
	Serial   string   `json:"serial"`
	RelayIDs []string `json:"rlys"`
}
