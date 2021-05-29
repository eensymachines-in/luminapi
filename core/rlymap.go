package core

/**/
import (
	"fmt"
	"reflect"
	"strings"
)

/*Relays are indentified by their IDS - IN1, IN2 IN3 ..
but this is only device compatible format. We cannot have user distinguish them from their IDs
Hence requires a Key-Map object to relate the ids and human readable description */
// ++++++++++++++++++++++++++++++++++++++++++++++ Relay Map++++++++++++++++++++++++++++
// IRelayMap :typical key:map object
type IRelayMap interface {
	AsMap() map[string]string
	IsInvalid() bool
	IsEqual(irm IRelayMap) bool
	KeyVal() string //gets the key of the map
	MapVal() string // gets the map value as string
}

type CollIRelayMap []IRelayMap

// CastEachTo : slice of Interfaces being cast to specific ojects here
func (crmp CollIRelayMap) CastEachTo(typ interface{}, cast func(IRelayMap) func() IRelayMap) interface{} {
	rtyp := reflect.TypeOf(typ)
	// https://tour.golang.org/moretypes/11
	// Learn about length and capacity of slices in GO
	// Here we want the slice to contain MAX len(crmp)
	// to start with it should have 0 elements , hence the arguments to MakeSlice
	result := reflect.MakeSlice(rtyp, 0, len(crmp))
	for _, item := range crmp {
		result = reflect.Append(result, reflect.ValueOf(cast(item)()))
	}
	return result.Interface()
}

// DevRelayMap : while devices can identify any relay by just their ids, human readable formats are slightly verbose
// Any device has a slice of these on the database, One RelayMap per relay [{IN1:"front door"}, {"IN2":"Floor1 & porch"}]
type DevRelayMap struct {
	Key string `json:"devrly" bson:"devrly"` // this value is the one that device understands
	Map string `json:"rlymap" bson:"rlymap"` // this value is human relatable: Floor1, Dining room, kitchen, bldg1-floor1
}

// CastFromIRelayMap : everytime you call this it sends out a function that can give you *devrelaymap
func (drm *DevRelayMap) CastFromIRelayMap(irmp IRelayMap) func() IRelayMap {
	return func() IRelayMap {
		return &DevRelayMap{Key: irmp.KeyVal(), Map: irmp.MapVal()}
	}
}

// Implementations on the DevRelayMap
func (drm *DevRelayMap) AsMap() map[string]string {
	result := map[string]string{}
	result[drm.Key] = drm.Map
	return result
}

func (drm *DevRelayMap) IsInvalid() bool {
	// definite and non empty key for the map makes it valid
	// key of the map to be non empty and less than 128
	return drm.Key == "" || len(drm.Key) > 128
}

func (drm *DevRelayMap) IsEqual(irm IRelayMap) bool {
	// 2 maps are equal when the relay ids are identical
	return strings.EqualFold(drm.Key, irm.KeyVal())
}

// Below are simple field properties
func (drm *DevRelayMap) KeyVal() string {
	return drm.Key
}
func (drm *DevRelayMap) MapVal() string {
	return drm.Map
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// Relay maps are found 2 forms - []IRelayMap , []string where []string is just the relay ids as the device understards
// map description of th ids is for human readable format
func RelayIdsFromMaps(maps []IRelayMap, result *[]string) error {
	*result = []string{}
	for _, rm := range maps {
		*result = append(*result, rm.KeyVal())
	}
	return nil
}

// takes 2 distinct slices of strings and makes an array of relay maps
func RelayMapsFromIds(ids, maps []string, result *[]IRelayMap) error {
	if len(ids) < len(maps) {
		return fmt.Errorf("have to be atleast as much ids as maps, cannot be lesser")
	}
	for i, id := range ids {
		*result = append(*result, &DevRelayMap{Key: id, Map: maps[i]})
	}
	return nil
}
