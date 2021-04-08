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

const (
	bUrl = "http://localhost/api/v1/devices"
)

// DeviceResponse :extension of the http response so that we can add some functions onto it
type DeviceResponse struct {
	*http.Response
}

/*Closure that sends abck a test func. Context of response and the status selection as params
send the http.Response from the calling test function to this, and a filter function that can pick the status code as required
Not all status codes shall warrant the reading of the response body*/
func ReadBody(resp *http.Response, status_allowed func(int) bool) func(*testing.T) {
	return func(t *testing.T) {
		if status_allowed(resp.StatusCode) {
			// If the request succeeds
			defer resp.Body.Close()
			target := []*scheduling.JSONRelayState{}
			if json.NewDecoder(resp.Body).Decode(&target) != nil {
				t.Log("We had a problem reading the response body - json.NewDecoder.Decode")
			}
			for _, sched := range target {
				t.Log(sched)
			}
		}
	}
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
	ReadBody(resp, func(code int) bool {
		return code == 200
	})(t)
	// When the device is already regsitered, re attempting will yeild 200 ok - quickly
	// The actual request
	resp, err = (&http.Client{}).Do(MakeNewDevicePOSTReq(serial, relays))
	// testing
	assert.Nil(t, err, "Unexpected error making a get request")
	assert.NotNil(t, resp, "Unexpected nil response from server")
	assert.Equal(t, resp.StatusCode, 200, "Unexpected status code in http response")

	// trying to GET the schedules for a registered devices
	t.Log("Now trying to get the device schedules")
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s", bUrl, serial), nil)
	resp, err = (&http.Client{}).Do(req)
	assert.Nil(t, err, "Unexpected error making a get request")
	assert.NotNil(t, resp, "Unexpected nil response from server")
	assert.Equal(t, resp.StatusCode, 200, "Unexpected status code in http response")

	t.Log("Now testing with nil payload on posting new device registration")
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/", bUrl), bytes.NewBuffer(nil))
	resp, err = (&http.Client{}).Do(req)
	assert.Nil(t, err, "Unexpected error making a get request")
	assert.Equal(t, 400, resp.StatusCode, "Unexpected status code in http response")

	// Now sending a rotten payload - check to see if the api can reject that with appropriate code
	t.Log("Now trying to register a device with invalid serial number")
	resp, err = (&http.Client{}).Do(MakeNewDevicePOSTReq("", relays))
	assert.Nil(t, err, "Unexpected error making a get request")
	assert.Equal(t, 400, resp.StatusCode, "Unexpected status code in http response")

	t.Log("Now trying to register a device with empty relay ids")
	// Please see : here we cannot use the same serial number, since that is registered and woudl return 200 OK without checking to see if the relays invliad
	resp, err = (&http.Client{}).Do(MakeNewDevicePOSTReq("random4w42", []string{}))
	assert.Nil(t, err, "Unexpected error making a get request")
	assert.Equal(t, 400, resp.StatusCode, "Unexpected status code in http response")

	// Getting the mqtt subscriber ready

	// Here we attempt to patch the device's schedule - this shall replace the schedules on the device
	// former schedules will be patched
	t.Log("Now trying to patch the schedules for the device")
	newScheds := []scheduling.JSONRelayState{
		{ON: "06:00 PM", OFF: "06:00 AM", IDs: []string{"IN1", "IN2", "IN3", "IN4"}, Primary: true},
		{ON: "06:30 PM", OFF: "06:01 PM", IDs: []string{"IN1", "IN2", "IN3", "IN4"}, Primary: false},
	}
	body, _ := json.Marshal(newScheds)
	req, _ = http.NewRequest("PATCH", fmt.Sprintf("%s/%s", bUrl, serial), bytes.NewBuffer(body))
	resp, err = (&http.Client{}).Do(req)
	assert.Nil(t, err, "Unexpected error making a get request")
	assert.Equal(t, 200, resp.StatusCode, "Unexpected status code in http response")

	/*now trying to patch schedules that have conflicts and check the response from api*/
	conflicScheds := []scheduling.JSONRelayState{
		{ON: "06:00 PM", OFF: "06:00 AM", IDs: []string{"IN1", "IN2", "IN3", "IN4"}, Primary: true},
		{ON: "05:50 PM", OFF: "06:30 PM", IDs: []string{"IN1", "IN2", "IN3", "IN4"}, Primary: false},
	}
	body, _ = json.Marshal(conflicScheds)
	req, _ = http.NewRequest("PATCH", fmt.Sprintf("%s/%s", bUrl, serial), bytes.NewBuffer(body))
	resp, err = (&http.Client{}).Do(req)
	assert.Nil(t, err, "Unexpected error making a get request")
	assert.Equal(t, 400, resp.StatusCode, "Unexpected status code in http response")
	ReadBody(resp, func(code int) bool {
		return code == 400 || code == 200
	})(t)

	// Now trying to patch the schedules of a device that does not exists
	t.Log("Now trying to patch the schedules for the device that isnt registered")
	req, _ = http.NewRequest("PATCH", fmt.Sprintf("%s/%s", bUrl, "somerandom4543"), bytes.NewBuffer(body))
	resp, err = (&http.Client{}).Do(req)
	assert.Nil(t, err, "Unexpected error making a patch request")
	assert.Equal(t, 404, resp.StatusCode, "Unexpected status code in http response")

	// // Now sending empty schedule pack to the device - this should be allowed
	t.Log("Now trying to patch empty schedule pack")
	body, _ = json.Marshal([]scheduling.JSONRelayState{}) //empty schedule pack
	req, _ = http.NewRequest("PATCH", fmt.Sprintf("%s/%s", bUrl, serial), bytes.NewBuffer(body))
	resp, err = (&http.Client{}).Do(req)
	assert.Nil(t, err, "Unexpected error making a patch request")
	assert.Equal(t, 200, resp.StatusCode, "Unexpected status code in http response")

	// Now here we are trying to remove a device registration
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("%s/%s", bUrl, serial), nil)
	resp, err = (&http.Client{}).Do(req)
	assert.Nil(t, err, "Unexpected error making a get request")
	assert.Equal(t, 200, resp.StatusCode, "Unexpected status code in http response")

	// Then trying again to delete the same serial device will fail
	resp, err = (&http.Client{}).Do(req)
	assert.Nil(t, err, "Unexpected error making a get request")
	assert.Equal(t, 404, resp.StatusCode, "Unexpected status code in http response")
}
