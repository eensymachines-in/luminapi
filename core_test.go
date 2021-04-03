package main

import (
	"testing"

	"github.com/eensymachines-in/luminapi/core"
	"github.com/eensymachines-in/scheduling"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

func TestCoreRegister(t *testing.T) {
	log.Debug("------------ Testing core device registry functions on luminapi-----------")
	// this shall test all the dev registry functions
	session, err := mgo.Dial("localhost:37017")
	if err != nil {
		return
	}
	log.WithFields(log.Fields{
		"address": "localhost:37017",
	}).Info("We now have a connection to the database")
	defer session.Close()
	// struting up the data
	coll := &core.DevRegsColl{Collection: session.DB("lumin").C("devreg")}
	serial := "851444e5"
	reg := &core.DevReg{}
	log.Info("Now checking to see if the device is registered")

	yes, err := coll.IsRegistered(serial)
	assert.Nil(t, err, "Wasnt expecting error in querying registery of the device")
	assert.Equal(t, yes, false, "the device wasnt expected to be registered")

	if !yes {
		// In all sanity we wouldnt want the device to be re registered
		log.Info("Now registering a new device")
		// finally registering the device
		err = coll.Register(serial, []string{"IN1", "IN2", "IN3", "IN4"}, reg)
		assert.Nil(t, err, "Unexpected error when registering a new device")
		if err != nil {
			return
		}
		// now testing to see if we can get the schedules from the registered device
		schedules := []*scheduling.JSONRelayState{}
		err = coll.GetSchedules(serial, &schedules)
		assert.Nil(t, err, "Was expecting a error when getting schedules of registered devices")
		if err != nil {
			return
		}
		log.WithFields(log.Fields{
			"schedules": schedules,
		}).Info("We are able to retrieve schedules for registered device")

		// With that its time now to see if we can update the device schedules
		schedules[0].OFF = "07:30 AM"
		schedules[0].ON = "05:22 PM"
		err = coll.UpdateSchedules(serial, schedules)
		assert.Nil(t, err, "Was expecting a error updating device's schedule")
		if err != nil {
			return
		}
		// Checking to see registry of the device.
		// the device now should be registered
		yes, err = coll.IsRegistered(serial)
		assert.Nil(t, err, "Wasnt expecting error in querying registery of the device")
		assert.Equal(t, yes, true, "The device was expected to be registered")
		if yes {
			err = coll.UnRegister(serial)
			assert.Nil(t, err, "Unexpected error in unregistering the device")
		}
	}

}
