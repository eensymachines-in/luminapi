/*A package that defines the core buisness logic for luminapi
A device when registered with the database, stores away the schedules against a unique serial number of the device.This database forms the single source of truth - modified by the user and followed by the device. Core package helps to CRUD this same datamodel. This package is expected to be used by api handlers ontop*/
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

// extension of the mgo collection - Make an object of this and then call functions specific so as to perform database CRUD operations
type DevRegsColl struct {
	*mgo.Collection
}

// GETs the schedules for the device serial
// Filters the device using the serial, sendsback a slice of scheduling.JSONRelayState, Errors when device is not registered or query fails
/*	result := []*scheduling.JSONRelayState{}
	serial := 'random45354'
	if err:=drc.GetSchedules(serial, &result); err! =nil{
		log.Errorf("Failed to get schedules for the device %s: %s", serial, err)
	}
	for _,s := range result{
		log.Info(*s)
	}
*/
func (drc *DevRegsColl) GetSchedules(serial string, result *[]scheduling.JSONRelayState) error {
	yes, err := drc.IsRegistered(serial)
	if err != nil {
		return errx.NewErr(&errx.ErrQuery{}, err, "Failed to check if the device is registered", "DevRegsColl/GetSchedules/IsRegistered")
	}
	if !yes {
		return errx.NewErr(&errx.ErrNotFound{}, err, "No schedules for unregistered devices", "DevRegsColl/GetSchedules")
	}
	reg := &DevReg{}
	if drc.Find(bson.M{"serial": serial}).One(reg) != nil {
		return errx.NewErr(&errx.ErrQuery{}, err, "Failed operation to get device schedules", "DevRegsColl/GetSchedules/One")
	}
	*result = reg.Schedules
	return nil
}

//Can just verify if the device is registered, errors when the database query fails
func (drc *DevRegsColl) IsRegistered(serial string) (bool, error) {
	count, err := drc.Find(bson.M{"serial": serial}).Count()
	if err != nil {
		return false, errx.NewErr(&errx.ErrQuery{}, err, "Failed to check if the device is registered", "DevRegsColl/IsRegistered")
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

// Registers a new device and sends back the default schedule
// error in case the register was unsuccessful or the inputs are invalid
// Inputs required are serial of the device and the relay ids
func (drc *DevRegsColl) Register(serial string, rlyIDS []string, result *DevReg) error {
	if serial == "" || len(rlyIDS) == 0 {
		// simple error check - 400 bad request
		return errx.NewErr(&errx.ErrInvalid{}, nil, "Serial of devices is invalid, or there arent enough relay definitions", "DevRegsColl")
	}
	// When the server sets it would add the defaul timing to the relay ids the client tells it to
	// When defaulting the server will not add any patch schedules
	*result = DevReg{Serial: serial, Schedules: []scheduling.JSONRelayState{
		{ON: "06:30 PM", OFF: "06:30 AM", IDs: rlyIDS, Primary: true},
	}}
	// default schedule gets pushed to the collection
	if err := drc.Insert(result); err != nil {
		return errx.NewErr(&errx.ErrQuery{}, err, "Failed operation to add device registration", "DevRegsColl/Register")
	}
	return nil
}

// Shall remove the device's entry from the database with all its schedules
func (drc *DevRegsColl) UnRegister(serial string) error {
	if err := drc.Remove(bson.M{"serial": serial}); err != nil {
		return errx.NewErr(&errx.ErrQuery{}, err, "Failed operation to unregister device", "DevRegsColl/UnRegister")
	}
	return nil
}

// Shall replace the schedules with new set of schedules, this works for one device serial
// makes no changes to schedules, just replaces them with a new set of schedules
func (drc *DevRegsColl) UpdateSchedules(serial string, newScheds []scheduling.JSONRelayState) error {
	yes, err := drc.IsRegistered(serial)
	if err != nil {
		return err
	}
	if !yes {
		return errx.NewErr(&errx.ErrNotFound{}, err, fmt.Sprintf("Failed to update schedule, since device %s is not registered", serial), "DevRegsColl/UpdateSchedules")
	}
	if err := drc.Update(bson.M{"serial": serial}, bson.M{"$set": bson.M{"schedules": newScheds}}); err != nil {
		return errx.NewErr(&errx.ErrQuery{}, err, "Failed operation to update device schedules", "DevRegsColl/UpdateSchedules")
	}
	return nil
}
