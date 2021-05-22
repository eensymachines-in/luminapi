package core

import "github.com/eensymachines-in/scheduling"

//datamodel of the device registration in the database
type DevReg struct {
	Serial    string                      `json:"serial" bson:"serial"` // unique serial number of the device
	Schedules []scheduling.JSONRelayState `json:"schedules" bson:"schedules"`
}

// +++++++++++++++++++++ Interfaces ++++++++++++++++++++
type Payload interface {
	Serial() string
}

// LogPayload : specific to DevLogPayload
type LogPayload interface {
	Logs() []map[string]interface{}
}

// RegPayload : specific for DevRegHttpPayload
type RegPayload interface {
	Relays() []string
}

// ++++++++++++++++++++++++++++++++++++++++++++++

// ++++++++++++++ Implementations +++++++++++++++++
// DevLogPayload : when the client intends to send the log dump to the api
type DevLogPayload struct {
	SerialID string                   `json:"serial"`
	LogData  []map[string]interface{} `json:"logdata"`
}

func (dlp *DevLogPayload) Serial() string {
	return dlp.SerialID
}
func (dlp *DevLogPayload) Logs() []map[string]interface{} {
	return dlp.LogData
}

// this one is the device registration as a simple http payload incoming
// When registering a new device all that we need is a serial number and the list of relay ids. Use this as a vehicle to unmarshal json objects
type DevRegHttpPayload struct {
	SerialID string   `json:"serial"`
	RelayIDs []string `json:"rlys"`
}

func (drp *DevRegHttpPayload) Serial() string {
	return drp.SerialID
}

func (drp *DevRegHttpPayload) Relays() []string {
	return drp.RelayIDs
}

// ++++++++++++++++++++++++++++++++++++++++++++++
