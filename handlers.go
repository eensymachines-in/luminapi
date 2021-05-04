package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/eensymachines-in/errx"
	"github.com/eensymachines-in/luminapi/core"
	"github.com/eensymachines-in/scheduling"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// HndlLogs : in the context of the log file path this send out a handler used by the api to output logs
func HndlLogs(filePath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			file, err := os.Open(filePath)
			if err != nil {
				// failed to open the log path file, hence sending back an error
				errx.DigestErr(errx.NewErr(errx.ErrInvalid{}, nil, "Failed to get log file at path. Check to see if logging is set", "HndlLogs/os.Open"), c)
				return
			}
			defer file.Close()
			reader := bufio.NewReader(file)
			result := []string{}
			for i := 0; i < 100; i++ {
				l, _, _ := reader.ReadLine()
				if len(string(l)) == 0 {
					// this generally means we have reached the end of the file
					break
				}
				result = append(result, string(l))
			}
			c.JSON(http.StatusOK, result)
		}
	}
}

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
		payload := []scheduling.JSONRelayState{}
		if c.ShouldBindJSON(&payload) != nil {
			errx.DigestErr(errx.NewErr(errx.ErrJSONBind{}, nil, "failed to read new schedules", "HandlDevice/ShouldBindJSON"), c)
			return
		}
		// ++++++++++++ here we check for any conflicts within the schedules
		sojrs := scheduling.SliceOfJSONRelayState(payload)
		scheds := []scheduling.Schedule{}
		sojrs.ToSchedules(&scheds)
		conflicting := []scheduling.JSONRelayState{} // from the payload this will get the conflicting ones
		for i, s := range scheds {
			if s.Conflicts() > 0 {
				// atleast one of the schedules has conflicts
				// we will accumulate the conflicting schdules in on temp array
				conflicting = append(conflicting, payload[i])
			}
		}
		// accumulated conflicts are then sent back as payload
		// TODO: test here from the client
		if len(conflicting) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message":   "One or more schedules have conflicts",
				"conflicts": conflicting,
			})
			log.WithFields(log.Fields{
				"conflicting": conflicting,
			}).Warn("We have atleast one schedule that has conflicts")
			return
		}
		// ++++++++++++ all the below code will run only if there arent any conflicting schedules
		// Conflicts if found then would send back the schedules as is ErrInvalid
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
		result := []scheduling.JSONRelayState{}
		if errx.DigestErr(devreg.GetSchedules(serial, &result), c) != 0 {
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}
}
