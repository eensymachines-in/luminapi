package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func TestDevReg(t *testing.T) {
	// getting a default devreg
	relaymaps := []IRelayMap{
		&RelayMap{RID: "IN1", Defn: "Corridor, front porch and basement"},
		&RelayMap{RID: "IN2", Defn: "Floor-1,4,7,11"},
		&RelayMap{RID: "IN3", Defn: "Floor-2,5,8,12"},
		&RelayMap{RID: "IN4", Defn: "Floor-3,6,9,13"},
	}
	reg := IReg(NewDevReg("564rfgfh", relaymaps))
	assert.NotNil(t, reg, "Unexpected nil dev registration : Failed NewDevReg")
	t.Log(reg)
	// Getting the serial of the device registration
	t.Log(reg.Serial())
	t.Log(reg.RegAsJsonStr())
	// Trying out non query log functions
	t.Log(reg.(ILogs).LogData())
	t.Log(reg.(IScheds).Schedules())
	t.Log(reg.(IRMaps).RelayIds())
	t.Log(reg.(IRMaps).RelayMaps())

	// Now testing queries
	t.Log("Inserting a new device registration")
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:   []string{"localhost:37017"},
		Timeout: 3 * time.Second,
	})
	if err != nil {
		t.Errorf("Failed to connect to localdatabase %s", err)
	}
	coll := session.DB("test").C("devreg")
	assert.Nil(t, coll.Insert(reg), "Unexpected error when inserting registration to the database")
	result := &DevReg{}
	reg.OfSerial(func(flt bson.M) error {
		coll.Find(flt).One(&result)
		t.Log("Now logging the result of Find Query ..")
		t.Log(result)
		return nil
	})
	// HEre we try to push logs into the registration
	dreg := &DevReg{
		SID: reg.Serial(),
		LData: []map[string]interface{}{
			{"level": "info", "msg": "Verbose logging", "time": "2021-03-16T09:07:57+05:30"},
			{"flog": true, "level": "info", "msg": "Starting autolumin module", "time": "2021-03-16T09:07:57+05:30", "verbose": true},
			{"level": "info", "msg": "Relayboard server: initializing...", "time": "2021-03-16T09:07:57+05:30"},
			{"level": "info", "msg": "Starting SrvRelay..", "status": map[string]interface{}{"IN1": 0, "IN2": 0, "IN3": 0, "IN4": 0}, "time": "2021-05-16T09:07:57+05:30"},
		},
	}
	t.Log("Now pushing logs to the registration")
	ILogs(dreg).QPushLogs(func(sel, upd bson.M) error {
		assert.Nil(t, coll.Update(sel, upd), "Unexpected error when pushing logs: coll.Update")
		return nil
	})
	t.Log("Now trying to get recent logs")
	dreg = &DevReg{
		SID: reg.Serial(),
	}
	// recentLogs := []struct {
	// 	Serial string                   `bson:"serial"`
	// 	Logs   []map[string]interface{} `bson:"logs"`
	// }{}
	recentLogs := []bson.M{}
	err = ILogs(dreg).QRecentLogs(-1, func(m []bson.M) error {
		t.Log(m)
		assert.Nil(t, coll.Pipe(m).All(&recentLogs), "Unexpected error when getting recent logs")
		t.Log(recentLogs)
		return nil
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Now outputting the recent log for the device")

	// Clearing the database of the setup registration
	reg.OfSerial(func(sel bson.M) error {
		assert.Nil(t, coll.Remove(sel), "Unexpected  error when removing registration from the database")
		return nil
	})

}
