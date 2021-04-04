package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/eensymachines-in/luminapi/core"
	"github.com/eensymachines-in/scheduling"
	"github.com/stretchr/testify/assert"
)

// DeviceResponse :extension of the http response so that we can add some functions onto it
type DeviceResponse struct {
	*http.Response
}

// ReadOut : this will read out the response payload of schedules
func (dr *DeviceResponse) ReadOut(t *testing.T) {
	// Investinggating the response body
	if dr.StatusCode == 200 {
		// If the request succeeds
		defer dr.Body.Close()
		target := []*scheduling.JSONRelayState{}
		if json.NewDecoder(dr.Body).Decode(&target) != nil {
			t.Log("We had a problem reading the response body - json.NewDecoder.Decode")
		}
		for _, sched := range target {
			t.Log(sched)
		}
	}
}

// MakeNewDevicePOSTReq : a fucntion that takes in the serial and the list of relays, bakes up a request and sends back to the testing function
func MakeNewDevicePOSTReq(s string, rlys []string) *http.Request {
	bUrl := "http://localhost/devices"
	// payload thats expected in the request
	payload := &core.DevRegHttpPayload{
		Serial:   s,
		RelayIDs: rlys,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/", bUrl), bytes.NewBuffer(body))
	return req
}
func TestDevices(t *testing.T) {
	// test data
	serial := "bb310d8176a2"
	relays := []string{"IN1", "IN2", "IN3", "IN4"}
	// The actual request
	resp, err := (&http.Client{}).Do(MakeNewDevicePOSTReq(serial, relays))
	// testing
	assert.Nil(t, err, "Unexpected error making a get request")
	assert.NotNil(t, resp, "Unexpected nil response from server")
	assert.Equal(t, resp.StatusCode, 200, "Unexpected status code in http response")
	(&DeviceResponse{resp}).ReadOut(t)
	// When the device is already regsitered, re attempting will yeild 200 ok - quickly
	// The actual request
	resp, err = (&http.Client{}).Do(MakeNewDevicePOSTReq(serial, relays))
	// testing
	assert.Nil(t, err, "Unexpected error making a get request")
	assert.NotNil(t, resp, "Unexpected nil response from server")
	assert.Equal(t, resp.StatusCode, 200, "Unexpected status code in http response")
	// No need to read out the response body since it would an empty schedule slice
}
