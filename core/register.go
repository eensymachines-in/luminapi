/*A package that defines the core buisness logic for luminapi
A device when registered with the database, stores away the schedules against a unique serial number of the device.This database forms the single source of truth - modified by the user and followed by the device. Core package helps to CRUD this same datamodel. This package is expected to be used by api handlers ontop*/
package core

import (
	"fmt"
	"time"

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

func (drc *DevRegsColl) GetDeviceLogs(serial string, flt string, result *[]map[string]string) error {
	// Checking to if the device is registered
	yes, err := drc.IsRegistered(serial)
	if err != nil {
		return errx.NewErr(&errx.ErrQuery{}, err, "Failed to check if the device is registered", "DevRegsColl/AppendDeviceLog/IsRegistered")
	}
	if !yes {
		return errx.NewErr(&errx.ErrNotFound{}, err, "No schedules for unregistered devices", "DevRegsColl/GetSchedules")
	}
	// Query to get the logs of device
	match_serial := bson.M{"$match": bson.M{"serial": serial}}
	stage_unwind := bson.M{"$unwind": bson.M{"path": "$logs"}}
	stage_project := bson.M{"$project": bson.M{"logs": 1, "serial": 1, "_id": 0}}
	sort_time := bson.M{"$sort": bson.M{"time": 1}} // time sorted logs
	matchQ := bson.M{}
	if flt != "" {
		matchQ = bson.M{"logs.level": flt}
	}
	match_lvl := bson.M{"$match": matchQ} // filter on the level of log
	qResult := []struct {
		Logs map[string]string `bson:"logs"`
	}{}
	if err := drc.Pipe([]bson.M{match_serial, stage_unwind, stage_project, sort_time, match_lvl}).All(&qResult); err != nil {
		return errx.NewErr(&errx.ErrQuery{}, err, fmt.Sprintf("Failed to get device logs for  %s", serial), "DevRegsColl/GetDeviceLogs/Pipe")
	}
	*result = []map[string]string{} // empty result
	for _, l := range qResult {
		*result = append(*result, l.Logs) //transfering the log items to the result
	}
	return nil
}

// AppendDeviceLog : appends the log blob to device registration
func (drc *DevRegsColl) AppendDeviceLog(serial string, logs []map[string]interface{}) error {
	yes, err := drc.IsRegistered(serial)
	if err != nil {
		return errx.NewErr(&errx.ErrQuery{}, err, "Failed to check if the device is registered", "DevRegsColl/AppendDeviceLog/IsRegistered")
	}
	if !yes {
		return errx.NewErr(&errx.ErrNotFound{}, err, "No schedules for unregistered devices", "DevRegsColl/GetSchedules")
	}
	// logs can quickly build up space and hence can be taxing on cloud machine
	// given that we want multiple devices as such we need to figure out a way to clear the logs as and when the device tries to push more
	// each time the device uploads the data, old logs can be cleared out
	match_serial := bson.M{"$match": bson.M{"serial": serial}}
	stage_unwind := bson.M{"$unwind": bson.M{"path": "$logs"}}
	stage_project := bson.M{"$project": bson.M{"logs": 1, "serial": 1, "_id": 0}}
	// at this stage we have all the logs for a single device unwound
	// result also has only select fields in them
	// RFC3339 is compatible format with mongo queries
	time_limit := bson.M{"$match": bson.M{"logs.time": bson.M{"$gte": time.Now().AddDate(0, -1, 0).Format(time.RFC3339)}}}
	// this will induce a time ceiling, all logs of a month always in the database
	sort_time := bson.M{"$sort": bson.M{"time": 1}}
	grp_logs := bson.M{"$group": bson.M{"_id": "$serial", "logs": bson.M{"$push": "$logs"}}}
	// sorted ascending by time the logs are grouped back into an array
	result := []struct {
		Serial string              `bson:"_id"`
		Logs   []map[string]string `bson:"logs"`
	}{} //expected result data shape - since we have grp query, result expected is only one
	// Incase there are no logs for the device this statement would fail
	if drc.Pipe([]bson.M{match_serial, stage_unwind, stage_project, time_limit, sort_time, grp_logs}).All(&result) != nil {
		return errx.NewErr(&errx.ErrQuery{}, err, fmt.Sprintf("Failed to get old logs for the device %s", serial), "DevRegsColl/AppendDeviceLog/Pipe")
	}
	if len(result) > 0 {
		if drc.Update(bson.M{"serial": serial}, bson.M{"$set": bson.M{"logs": result[0].Logs}}) != nil {
			return errx.NewErr(&errx.ErrQuery{}, err, fmt.Sprintf("Failed to trim/limit logs for device %s", serial), "DevRegsColl/AppendDeviceLog/Update")
		}
	}
	// We then update the device registration for the logs
	if len(logs) > 0 {
		if drc.Update(bson.M{"serial": serial}, bson.M{"$push": bson.M{"logs": bson.M{"$each": logs}}}) != nil {
			return errx.NewErr(&errx.ErrQuery{}, err, fmt.Sprintf("Failed to push logs for device %s", serial), "DevRegsColl/AppendDeviceLog/Update")
		}
	}
	return nil
}
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
