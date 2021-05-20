package core

import (
	"testing"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func ConnectionOrPanic(coll string) *mgo.Collection {
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:   []string{"localhost:37017"},
		Timeout: 10 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	result := session.DB("lumin").C(coll)
	if result == nil {
		panic("Failed to get collection from database")
	}
	return result
}

func TestLogGet(t *testing.T) {
	coll := ConnectionOrPanic("devreg")
	match_serial := bson.M{"$match": bson.M{"serial": "000000007920365b"}}
	stage_unwind := bson.M{"$unwind": bson.M{"path": "$logs"}}
	stage_project := bson.M{"$project": bson.M{"logs": 1, "serial": 1, "_id": 0}}
	sort_time := bson.M{"$sort": bson.M{"time": 1}}
	match_lvl := bson.M{"$match": bson.M{"logs.level": "gdfgdfg"}}
	result := []struct {
		Logs map[string]string `bson:"logs"`
	}{}
	if err := coll.Pipe([]bson.M{match_serial, stage_unwind, stage_project, sort_time, match_lvl}).All(&result); err != nil {
		t.Error(err)
	}
	for _, r := range result {
		t.Log(r)
	}
}

func TestLogRemove(t *testing.T) {
	coll := ConnectionOrPanic("devreg")
	match_serial := bson.M{"$match": bson.M{"serial": "000000007920365b"}}
	stage_unwind := bson.M{"$unwind": bson.M{"path": "$logs"}}
	stage_project := bson.M{"$project": bson.M{"logs": 1, "serial": 1, "_id": 0}}
	time_limit := bson.M{"$match": bson.M{"logs.time": bson.M{"$gte": time.Now().AddDate(0, -1, 0).Format(time.RFC3339)}}}
	sort_time := bson.M{"$sort": bson.M{"time": 1}}
	grp_logs := bson.M{"$group": bson.M{"_id": "$serial", "logs": bson.M{"$push": "$logs"}}}
	result := struct {
		Serial string              `bson:"_id"`
		Logs   []map[string]string `bson:"logs"`
	}{}
	coll.Pipe([]bson.M{match_serial, stage_unwind, stage_project, time_limit, sort_time, grp_logs}).One(&result)
	t.Log(result)

}
