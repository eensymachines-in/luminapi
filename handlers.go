package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/eensymachines-in/errx"
	"github.com/eensymachines-in/luminapi/core"
	"github.com/eensymachines-in/scheduling"
	"github.com/gin-gonic/gin"
)

func HandlDevices(c *gin.Context) {
	// Getting values from middleware - database connection and clean up closures
	val, exists := c.Get("devreg")
	if !exists {
		errx.DigestErr(errx.NewErr(errx.ErrConnFailed{}, nil, "failed connection to database", "HandlDevices/dbConnect"), c)
		return
	}
	devreg, _ := val.(*core.DevRegsColl)
	val, _ = c.Get("db_close")
	dbClose := val.(func())
	defer dbClose()
	// +++++++++++++++++++++++++ method verb definitions ++++++++++++++++++
	if c.Request.Method == "POST" {
		// will check if the device is registered , if it already is then would return 200 ok
		// else shall register and then send ok
		val, _ = c.Get("devregpayload")
		payload, _ := val.(*core.DevRegHttpPayload)
		val, _ = c.Get("isreg")
		yes, _ := val.(bool)
		if yes {
			// Incase the device is already registered we just quit from this branch
			// no need to re-register the device
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		// this in case the device is not registered
		registration := &core.DevReg{}
		if errx.DigestErr(devreg.Register(payload.Serial, payload.RelayIDs, registration), c) != 0 {
			return
		}
		// when device registers itself newly the response also has schedules
		c.JSON(http.StatusOK, registration.Schedules)
		return
	}
}
func HandlDevice(c *gin.Context) {
	val, exists := c.Get("devreg")
	if !exists {
		errx.DigestErr(errx.NewErr(errx.ErrConnFailed{}, nil, "failed connection to database", "HandlDevices/dbConnect"), c)
		return
	}
	devreg, _ := val.(*core.DevRegsColl)
	val, _ = c.Get("db_close")
	dbClose := val.(func())
	defer dbClose()
	serial := c.Param("serial")

	if c.Request.Method == "DELETE" {
		if errx.DigestErr(devreg.UnRegister(serial), c) != 0 {
			return
		}
		c.JSON(http.StatusOK, gin.H{})
		return
	} else if c.Request.Method == "PATCH" {
		val, _ = c.Get("mqttclient")
		mqttClient, _ := val.(mqtt.Client)
		defer mqttClient.Disconnect(250) // this is important to dispose

		// read in the the schedules that need to be patched
		// If the device is not registered, the middleware will handle it
		payload := []*scheduling.JSONRelayState{}
		if c.ShouldBindJSON(&payload) != nil {
			errx.DigestErr(errx.NewErr(errx.ErrJSONBind{}, nil, "failed to read new schedules", "HandlDevice/ShouldBindJSON"), c)
			return
		}
		if errx.DigestErr(devreg.UpdateSchedules(serial, payload), c) != 0 {
			return
		}
		// marshal the json string - set the topic and then off it goes
		mqttText, err := json.Marshal(payload)
		if err != nil {
			errx.DigestErr(err, c)
		}
		token := mqttClient.Publish(fmt.Sprintf("%s/schedules", serial), 0, false, string(mqttText))
		token.Wait()
		// time.Sleep(1 * time.Second)
		c.JSON(http.StatusOK, gin.H{})
		return
	} else if c.Request.Method == "GET" {
		// Gets the schedules for a device given the serial of the device
		result := []*scheduling.JSONRelayState{}
		if errx.DigestErr(devreg.GetSchedules(serial, &result), c) != 0 {
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}
}
