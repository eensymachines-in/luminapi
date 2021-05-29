package core

import (
	"encoding/json"
	"time"

	"github.com/eensymachines-in/scheduling"
	"gopkg.in/mgo.v2/bson"
)

type IReg interface {
	// HElps to get registration as json string
	RegAsJsonStr() string
}

//datamodel of the device registration in the database - this is distinct from the format whats communicated over http
// device registration has schedules and relaymaps
// it also has log data - an array of free form logs from the device
type DevReg struct {
	SID    string                      `json:"serial" bson:"serial"` // unique serial number of the device
	Scheds []scheduling.JSONRelayState `json:"scheds,omitempty" bson:"scheds"`
	RMaps  []*DevRelayMap              `json:"rmaps,omitempty" bson:"rmaps"`
	LData  []map[string]interface{}    `json:"logs,omitempty" bson:"logs"`
}

// RegAsJsonStr : I dont see this as value function, but we have a  interface attached to this
func (dr *DevReg) RegAsJsonStr() string {
	byt, _ := json.Marshal(dr)
	return string(byt)
}

// OfSerial : for any given DevReg it makes a query that can run in context of mgo.Collection to get the Devreg
func (dreg *DevReg) OfSerial(q func(flt bson.M) error) error {
	return q(bson.M{"serial": dreg.SID})
}

// GetLogs : prepares the query in the context of DevReg then calls the query function
// query function will run in the context of the then database and collection
func (dreg *DevReg) GetLogs(lvl string, q func(pipe []bson.M) error) error {
	match_serial := bson.M{"$match": bson.M{"serial": dreg.SID}}
	stage_unwind := bson.M{"$unwind": bson.M{"path": "$logs"}}
	// Unwinding the logs for the serial - this will be in [{serial:"", logs:map[string]interface{}}...]
	stage_project := bson.M{"$project": bson.M{"logs": 1, "serial": 1, "_id": 0}}
	sort_time := bson.M{"$sort": bson.M{"time": 1}} // time sorted logs
	matchQ := bson.M{}
	if lvl != "" {
		// this filter on the level of the logs is optional
		matchQ = bson.M{"logs.level": lvl}
	}
	match_lvl := bson.M{"$match": matchQ} // filter on the level of log
	return q([]bson.M{match_serial, stage_unwind, stage_project, sort_time, match_lvl})
}

// OldLogs : gets for a specified time all the logs that are after the cutoff
// monthsback : negative integer in months that the 'recent' is to be defined for, -1 = 1 month old logs
func (dreg *DevReg) RecentLogs(monthsback int, q func([]bson.M) error) error {
	match_serial := bson.M{"$match": bson.M{"serial": dreg.SID}}
	stage_unwind := bson.M{"$unwind": bson.M{"path": "$logs"}}
	stage_project := bson.M{"$project": bson.M{"logs": 1, "serial": 1, "_id": 0}}
	// at this stage we have all the logs for a single device unwound
	// result also has only select fields in them
	// RFC3339 is compatible format with mongo queries
	time_limit := bson.M{"$match": bson.M{"logs.time": bson.M{"$gte": time.Now().AddDate(0, monthsback, 0).Format(time.RFC3339)}}}
	// this will induce a time ceiling, all logs of a month always in the database
	sort_time := bson.M{"$sort": bson.M{"time": 1}}
	grp_logs := bson.M{"$group": bson.M{"_id": "$serial", "logs": bson.M{"$push": "$logs"}}}
	return q([]bson.M{match_serial, stage_unwind, stage_project, time_limit, sort_time, grp_logs})
}

// ReplaceLogs : completely replaces the logs node on the device registration
// caution : this is not reversible
func (dreg *DevReg) ReplaceLogs(q func(sel bson.M, upd bson.M) error) error {
	return q(bson.M{"serial": dreg.SID}, bson.M{"$set": bson.M{"logs": dreg.LData}})
}

// PushLogs : this does not replace the logs but pushes new logs to the existing ones
func (dreg *DevReg) PushLogs(q func(sel bson.M, upd bson.M) error) error {
	return q(bson.M{"serial": dreg.SID}, bson.M{"$push": bson.M{"logs": bson.M{"$each": dreg.LData}}})
}

// ReplaceScheds : replaces the scheds node for the device
func (dreg *DevReg) ReplaceScheds(q func(sel bson.M, upd bson.M) error) error {
	return q(bson.M{"serial": dreg.SID}, bson.M{"$set": bson.M{"scheds": dreg.Scheds}})
}

func (dreg *DevReg) Defaults(rmaps []IRelayMap) {
	ids := []string{}
	RelayIdsFromMaps(rmaps, &ids)
	drmp := &DevRelayMap{}
	sdrmp := []*DevRelayMap{}
	// making one primary schedule of default time
	dreg.Scheds = []scheduling.JSONRelayState{{ON: "06:30 PM", OFF: "06:30 AM", IDs: ids, Primary: true}}
	dreg.RMaps = CollIRelayMap(rmaps).CastEachTo(sdrmp, drmp.CastFromIRelayMap).([]*DevRelayMap)
	dreg.LData = []map[string]interface{}{}
}
