package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	reset  = iota
	red    = iota
	green  = iota
	yellow = iota
	blue   = iota
	purple = iota
	cyan   = iota
	gray   = iota
	white  = iota
)

func ColorMessage(c int, message string) string {
	return fmt.Sprintf("\033[3%dm %s \033[0m", c, message)
}

func TestDevReg(t *testing.T) {
	// getting a default devreg
	relaymaps := []IRelayMap{
		&RelayMap{RID: "IN1", Defn: "Corridor, front porch and basement"},
		&RelayMap{RID: "IN2", Defn: "Floor-1,4,7,11"},
		&RelayMap{RID: "IN3", Defn: "Floor-2,5,8,12"},
		&RelayMap{RID: "IN4", Defn: "Floor-3,6,9,13"},
	}
	t.Log(ColorMessage(green, "Now making a device registration with default schedules.."))
	reg := IReg(NewDevReg("564rfgfh", relaymaps))
	assert.NotNil(t, reg, "Unexpected nil dev registration : Failed NewDevReg")

	t.Log(ColorMessage(green, "Now checking the device registration for field values"))
	t.Log(reg)
	t.Log(reg.Serial())
	t.Log(reg.RegAsJsonStr())
	// Trying out non query log functions
	t.Log(reg.(ILogs).LogData())
	t.Log(reg.(IScheds).Schedules())
	t.Log(reg.(IRMaps).RelayIds())
	t.Log(reg.(IRMaps).RelayMaps())

	// Now testing queries
	t.Log(ColorMessage(yellow, "Now connecting to local database"))
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:   []string{"localhost:37017"},
		Timeout: 3 * time.Second,
	})
	if err != nil {
		t.Errorf(ColorMessage(red, fmt.Sprintf("Failed to connect to localdatabase %s", err)))
	}
	coll := session.DB("test").C("devreg")
	defer session.Close() // like a responsible citizen we close the database connection once its used

	t.Log(ColorMessage(green, "Now inserting new device registration"))
	assert.Nil(t, coll.Insert(reg), "Unexpected error when inserting registration to the database")
	result := &DevReg{}
	err = reg.OfSerial(func(flt bson.M) error {
		return coll.Find(flt).One(&result)
	})
	if err != nil {
		t.Errorf(ColorMessage(red, fmt.Sprintf("Failed to get device reg with serial %s", err)))
	}
	t.Log(result)
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
		assert.Nil(t, coll.Pipe(m).All(&recentLogs), "Unexpected error when getting recent logs")
		return nil
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Now outputting the recent log for the device")
	t.Log(recentLogs)
	// Then we try to replace the logs of the device with a bunch of new logs
	dreg = &DevReg{
		SID: reg.Serial(),
		LData: []map[string]interface{}{
			{"level": "warning", "msg": "SrvRelay: Interrupt!, now quitting..", "time": "2021-05-16T09:08:18+05:30"},
			{"level": "warning", "msg": "Now shutting down the RelayBoardOLED ..", "time": "2021-05-16T09:08:18+05:30"},
			{"level": "warning", "msg": "Now shutting down the relay IN1 ..", "time": "2021-05-16T09:08:18+05:30"},
			{"level": "warning", "msg": "Now shutting down the relay IN2 ..", "time": "2021-05-16T09:08:18+05:30"},
			{"level": "warning", "msg": "Now shutting down the relay IN3 ..", "time": "2021-05-16T09:08:18+05:30"},
			{"level": "warning", "msg": "Now shutting down the relay IN4 ..", "time": "2021-05-16T09:08:18+05:30"},
			{"level": "warning", "msg": "Relayboard server: quitting...", "time": "2021-05-16T09:08:18+05:30"},
		},
	}
	err = ILogs(dreg).QReplaceLogs(func(sel, upd bson.M) error {
		assert.Nil(t, coll.Update(sel, upd), "Unexpected error when updating the logs..")
		return nil
	})
	if err != nil {
		t.Error(err)
		return
	}
	ILogs(dreg).QRecentLogs(-1, func(m []bson.M) error {
		assert.Nil(t, coll.Pipe(m).All(&recentLogs), "Unexpected error when getting recent logs")
		return nil
	})
	t.Log("Now outputting the recent logs after they have been replaced")
	t.Log(recentLogs)
	// Now getting filtered logs
	ILogs(dreg).QGetLogs("warning", func(pipe []bson.M) error {
		assert.Nil(t, coll.Pipe(pipe).All(&recentLogs), "Unexpected error when getting logs")
		return nil
	})
	t.Log(ColorMessage(yellow, "Now outputting all the warning logs"))
	t.Log(recentLogs)
	// Clearing the database of the setup registration
	reg.OfSerial(func(sel bson.M) error {
		assert.Nil(t, coll.Remove(sel), "Unexpected  error when removing registration from the database")
		return nil
	})

}
