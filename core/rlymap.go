package core

/**/
import (
	"strings"
)

/*Relays are indentified by their IDS - IN1, IN2 IN3 ..
but this is only device compatible format. We cannot have user distinguish them from their IDs
Hence requires a RID-Defn object to relate the ids and human readable description */
// ++++++++++++++++++++++++++++++++++++++++++++++ Relay Defn++++++++++++++++++++++++++++
// IRelayMap :typical key:map object
type IRelayMap interface {
	AsMap() map[string]string
	IsInvalid() bool
	IsEqual(irm IRelayMap) bool
	RlyID() string      //gets the key of the map
	Definition() string // gets the map value as string
}

// DevRelayMap : while devices can identify any relay by just their ids, human readable formats are slightly verbose
// Any device has a slice of these on the database, One RelayMap per relay [{IN1:"front door"}, {"IN2":"Floor1 & porch"}]
type RelayMap struct {
	RID  string `json:"rid" bson:"rid"`   // this value is the one that device understands
	Defn string `json:"defn" bson:"defn"` // this value is human relatable: Floor1, Dining room, kitchen, bldg1-floor1
}

// Implementations on the DevRelayMap
func (drm *RelayMap) AsMap() map[string]string {
	result := map[string]string{}
	result[drm.RID] = drm.Defn
	return result
}

func (drm *RelayMap) IsInvalid() bool {
	// definite and non empty key for the map makes it valid
	// key of the map to be non empty and less than 128
	return drm.RID == "" || len(drm.RID) > 128
}

func (drm *RelayMap) IsEqual(irm IRelayMap) bool {
	// 2 maps are equal when the relay ids are identical
	return strings.EqualFold(drm.RID, irm.RlyID())
}

// Below are simple field properties
func (drm *RelayMap) RlyID() string {
	return drm.RID
}
func (drm *RelayMap) Definition() string {
	return drm.Defn
}
