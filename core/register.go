package core

import (
	"fmt"

	"github.com/eensymachines-in/errx"
	"github.com/eensymachines-in/scheduling"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/* The device registers itself via the authapi microservice, and then proceeds itself to register here on luminapi
Here all that is required is the device UID and the schedules underneath it */

// DevReg : represents one device registration in the database
type DevReg struct {
	Serial    string                       `json:"serial" bson:"serial"` // unique serial number of the device
	Schedules []*scheduling.JSONRelayState `json:"schedules" bson:"schedules"`
}

// DevRegsColl : extension of the mgo collection - make this object and call the functions on it
type DevRegsColl struct {
	*mgo.Collection
}

// GetSchedules : gets the schedules for the device serial
func (drc *DevRegsColl) GetSchedules(serial string, result *[]*scheduling.JSONRelayState) error {
	yes, err := drc.IsRegistered(serial)
	if err != nil {
		return errx.NewErr(errx.ErrQuery{}, err, "Failed to check if the device is registered", "DevRegsColl/GetSchedules/IsRegistered")
	}
	if !yes {
		return errx.NewErr(errx.ErrNotFound{}, err, "No schedules for unregistered devices", "DevRegsColl/GetSchedules")
	}
	reg := &DevReg{}
	if drc.Find(bson.M{"serial": serial}).One(reg) != nil {
		return errx.NewErr(errx.ErrQuery{}, err, "Failed operation to get device schedules", "DevRegsColl/GetSchedules/One")
	}
	*result = reg.Schedules
	return nil
}

// IsRegistered : this can just verify if the device is registered
func (drc *DevRegsColl) IsRegistered(serial string) (bool, error) {
	count, err := drc.Find(bson.M{"serial": serial}).Count()
	if err != nil {
		return false, errx.NewErr(errx.ErrQuery{}, err, "Failed to check if the device is registered", "DevRegsColl/IsRegistered")
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

// Register : registers a new device and sends back the default schedule
// error in case the register was unsuccessful
func (drc *DevRegsColl) Register(serial string, rlyIDS []string, result *DevReg) error {
	// When the server sets it would add the defaul timing to the relay ids the client tells it to
	// When defaulting the server will not add any patch schedules
	*result = DevReg{Serial: serial, Schedules: []*scheduling.JSONRelayState{
		{ON: "06:30 PM", OFF: "06:30 AM", IDs: rlyIDS, Primary: true},
	}}
	// default schedule gets pushed to the collection
	if err := drc.Insert(result); err != nil {
		return errx.NewErr(errx.ErrQuery{}, err, "Failed operation to add device registration", "DevRegsColl/Register")
	}
	return nil
}

// UnRegister : this shall remove the device's entry from the database with all its schedules
func (drc *DevRegsColl) UnRegister(serial string) error {
	if err := drc.Remove(bson.M{"serial": serial}); err != nil {
		return errx.NewErr(errx.ErrQuery{}, err, "Failed operation to unregister device", "DevRegsColl/UnRegister")
	}
	return nil
}

// UpdateSchedules : this shall replace the schedules with new set of schedules,
// this works for one device serial
// makes no changes to schedules, just replaces them with a new set of schedules
func (drc *DevRegsColl) UpdateSchedules(serial string, newScheds []*scheduling.JSONRelayState) error {
	yes, err := drc.IsRegistered(serial)
	if err != nil {
		return err
	}
	if !yes {
		return errx.NewErr(errx.ErrNotFound{}, err, fmt.Sprintf("Failed to update schedule, since device %s is not registered", serial), "DevRegsColl/UpdateSchedules")
	}
	if err := drc.Update(bson.M{"serial": serial}, bson.M{"$set": bson.M{"schedules": newScheds}}); err != nil {
		return errx.NewErr(errx.ErrQuery{}, err, "Failed operation to update device schedules", "DevRegsColl/UpdateSchedules")
	}
	return nil
}
