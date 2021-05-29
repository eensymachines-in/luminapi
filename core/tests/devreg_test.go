package core

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/eensymachines-in/luminapi/core"
	"github.com/eensymachines-in/scheduling"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func TestDevReg(t *testing.T) {
	result := []core.IRelayMap{}
	core.RelayMapsFromIds([]string{"IN1", "IN2", "IN3", "IN4"}, []string{"First floor", "Second floor", "Porch", "Podium"}, &result)

	drmp := &core.DevRelayMap{}
	sdrmp := []*core.DevRelayMap{}
	dreg := core.DevReg{
		SID:    "534543jfgdgd",
		Scheds: []scheduling.JSONRelayState{},
		RMaps:  core.CollIRelayMap(result).CastEachTo(sdrmp, drmp.CastFromIRelayMap).([]*core.DevRelayMap),
		LData:  []map[string]interface{}{},
	}
	// connecting to the test database
	// we need to test if RMaps are injected to the database as expected
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:   []string{"localhost:37017"},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Errorf("Failed to connect to local testing database %s", err)
	}
	coll := session.DB("test").C("devreg")
	if coll == nil {
		t.Error("Failed to connect to test database test/relaymaps")
	}
	t.Log(dreg)
	err = coll.Insert(dreg)
	assert.Nil(t, err, "Unexpected error when inserting device registration")
	// // Now lets find out if the unregistration also works fine
	dreg2 := bson.M{}
	err = coll.Find(bson.M{}).One(&dreg2)
	assert.Nil(t, err, "Unexpected error when getting the device registration")
	t.Log(dreg2)
}

func TestJsonDevReg(t *testing.T) {
	result := []core.IRelayMap{}
	core.RelayMapsFromIds([]string{"IN1", "IN2", "IN3", "IN4"}, []string{"First floor", "Second floor", "Porch", "Podium"}, &result)

	drmp := &core.DevRelayMap{}
	sdrmp := []*core.DevRelayMap{}
	dreg := core.DevReg{
		SID: "534543jfgdgd",
		// Scheds: []scheduling.JSONRelayState{},
		RMaps: core.CollIRelayMap(result).CastEachTo(sdrmp, drmp.CastFromIRelayMap).([]*core.DevRelayMap),
		// LData:  []map[string]interface{}{},
	}
	byt, err := json.Marshal(dreg)
	assert.Nil(t, err, "Unexpected error when MArshalling registration")
	dreg2 := core.DevReg{}
	t.Log(string(byt))
	err = json.Unmarshal(byt, &dreg2)
	assert.Nil(t, err, "Unexpected error when Unmarhsalling registration")
	t.Log(dreg2)
}
