package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/eensymachines-in/errx"
	"github.com/eensymachines-in/scheduling"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func HndlCommands(c *gin.Context) {
	if c.Request.Method == "POST" {
		action := c.Query("action")
		serial := c.Param("serial") // we should ideally check for the device if its registered
		// but since this is just a mqtt post we are skipping that for now
		// worst case is that some miscreant can push excessive random post requests to make the mqtt broker busy
		val, _ := c.Get("mqttclient")
		mqttClient, _ := val.(mqtt.Client)
		defer mqttClient.Disconnect(250) // this is important to dispose
		if action == "" {
			errx.DigestErr(errx.NewErr(&errx.ErrInvalid{}, nil, "Failed to get any action on the command", "HndlCommands"), c)
			return
		}
		if action == "shutdown" {
			// We just post the command to the mqtt broker
			mqttText := "shutdown=now"
			token := mqttClient.Publish(fmt.Sprintf("%s/commands", serial), 0, false, mqttText)
			token.Wait()
			log.Infof("Sent command to %s: shutdown", serial)
			c.AbortWithStatus(http.StatusOK)
			return
			// yeah thats it, api will only shuttle a text message to the mqtt broker
		}
		errx.DigestErr(errx.NewErr(&errx.ErrInvalid{}, nil, "Invalid action in the command", "HndlCommands"), c)
		return
	}
}

// HndlLogs : in the context of the log file path this send out a handler used by the api to output logs
func HndlLogs(filePath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			file, err := os.Open(filePath)
			if err != nil {
				// failed to open the log path file, hence sending back an error
				errx.DigestErr(errx.NewErr(&errx.ErrInvalid{}, err, "Failed to get log file at path. Check to see if logging is set", "HndlLogs/os.Open"), c)
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
		errx.DigestErr(errx.NewErr(&errx.ErrConnFailed{}, nil, "failed connection to database", "HandlDevices/dbConnect"), c)
		return
	}
	devreg, _ := val.(*mgo.Collection)
	val, _ = c.Get("db_close")
	dbClose := val.(func())
	defer dbClose()
	// +++++++++++++++++++++++++ method verb definitions ++++++++++++++++++
	if c.Request.Method == "POST" {
		// will check if the device is registered , if it already is then would return 200 ok
		// else shall register and then send ok
		val, _ = c.Get("dev_payload")
		payload, _ := val.(*DevReg)
		val, _ = c.Get("isreg")
		yes, _ := val.(bool)
		if yes {
			// Incase the device is already registered we just quit from this branch
			// no need to re-register the device
			// Duplicate registration does not throw errors
			res := &DevReg{}
			err := IReg(payload).OfSerial(func(flt bson.M) error {
				return devreg.Find(flt).One(res)
			})
			if err != nil {
				errx.DigestErr(errx.NewErr(&errx.ErrQuery{}, err, "Failed to insert new registration", "HandlDevices/OfSerial"), c)
				return
			}
			log.WithFields(log.Fields{
				"serial": IReg(payload).Serial(),
			}).Info("Device is already registered")
			c.JSON(http.StatusOK, res)
			// when the device is registered - it emits the schedules for the existing device
			// this is useful when if client has made changes to the schedules when the device was turned off..
			// upon restart the device will still get the changed schedules. Server will remain the source of truth
			return
		}
		// this in case the device is not registered
		// we make default schedules for the device and push the registration in the database
		insertion := NewDevReg(payload.SID, IRMaps(payload).RelayMaps())
		if err := devreg.Insert(insertion); err != nil {
			// this is incase the registration fails
			errx.DigestErr(errx.NewErr(&errx.ErrQuery{}, err, "Failed to insert new registration", "HandlDevices/Insert"), c)
			return
		}
		// when device registers itself newly the response also has schedules
		// newly created device registration will also have default schedules
		// default schedules are sent back to the device which will start its job along with this
		// but since the schedules are in slices we here pack it up in devreg format - easier on the client side to treat it as map[string]interface{}
		result := &DevReg{SID: insertion.SID, Scheds: IScheds(insertion).RelayStates()}
		c.JSON(http.StatusOK, result)
		return
	}
}
func HandlDevice(c *gin.Context) {
	val, exists := c.Get("devreg")
	if !exists {
		errx.DigestErr(errx.NewErr(&errx.ErrConnFailed{}, nil, "failed connection to database", "HandlDevices/dbConnect"), c)
		return
	}
	devreg, _ := val.(*mgo.Collection)
	val, _ = c.Get("db_close")
	dbClose := val.(func())
	defer dbClose()
	serial := c.Param("serial")
	if c.Request.Method == "DELETE" {
		item := &DevReg{SID: serial}
		err := item.OfSerial(func(flt bson.M) error {
			return devreg.Remove(flt)
		})
		if err != nil {
			errx.DigestErr(errx.NewErr(&errx.ErrQuery{}, err, "Failed to remove device registration", "HandlDevices/Remove"), c)
			return
		}
		log.WithFields(log.Fields{
			"serial": serial,
		}).Warn("Device registration being dropped")
		c.JSON(http.StatusOK, gin.H{})
		return
	} else if c.Request.Method == "PATCH" {
		val, _ = c.Get("mqttclient")
		mqttClient, _ := val.(mqtt.Client)
		defer mqttClient.Disconnect(250) // this is important to dispose
		val, _ = c.Get("dev_payload")
		payload, _ := val.(*DevReg)
		if payload == nil {
			errx.DigestErr(errx.NewErr(&errx.ErrInvalid{}, nil, "Failed to read payload, {'serial', 'scheds'}", "HandlDevice/PATCH"), c)
			return
		}
		if IReg(payload).Serial() != serial {
			log.Warn("Mismatch in the device serial and the url param")
			errx.DigestErr(errx.NewErr(&errx.ErrInvalid{}, nil, "serial of the device mismatching with the url param", "HandlDevice/PATCH"), c)
			return
		}
		// ++++++++++++ here we check for any conflicts within the schedules
		sojrs := scheduling.SliceOfJSONRelayState(IScheds(payload).RelayStates())
		scheds := []scheduling.Schedule{}
		sojrs.ToSchedules(&scheds)
		conflicting := []scheduling.JSONRelayState{} // from the payload this will get the conflicting ones
		for i, s := range scheds {
			if s.Conflicts() > 0 {
				// atleast one of the schedules has conflicts
				// we will accumulate the conflicting schdules in on temp array
				conflicting = append(conflicting, sojrs[i])
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
		err := IScheds(payload).QReplaceScheds(func(sel, upd bson.M) error {
			return devreg.Update(sel, upd)
		})
		if err != nil {
			errx.DigestErr(errx.NewErr(&errx.ErrQuery{}, err, "Failed to update schedules for device", "HandlDevice/QReplaceScheds"), c)
			return
		}
		// marshal the json string - set the topic and then off it goes
		mqttText, err := json.Marshal(IScheds(payload).RelayStates())
		if err != nil {
			errx.DigestErr(err, c)
		}
		token := mqttClient.Publish(fmt.Sprintf("%s/schedules", serial), 0, false, string(mqttText))
		token.Wait()
		log.Info("Now returning 200 ok with gin empty result ..")
		c.JSON(http.StatusOK, gin.H{})
		return
	} else if c.Request.Method == "GET" {
		// Gets the schedules for a device given the serial of the device
		result := &DevReg{}
		err := (&DevReg{SID: serial}).OfSerial(func(flt bson.M) error {
			return devreg.Find(flt).One(result)
		})
		if err != nil {
			errx.DigestErr(errx.NewErr(&errx.ErrQuery{}, err, "Failed to get schedules for device", "HandlDevice/OfSerial"), c)
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}
}

// HndlDeviceLogs: can handle logs CRUD specific to the device
func HndlDeviceLogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Getting the database connection
		val, exists := c.Get("devreg")
		if !exists {
			errx.DigestErr(errx.NewErr(&errx.ErrConnFailed{}, nil, "failed connection to database", "HandlDevices/dbConnect"), c)
			return
		}
		devreg, _ := val.(*mgo.Collection)
		val, _ = c.Get("db_close")
		dbClose := val.(func())
		defer dbClose()
		// Now that we get the payload bound from middleware
		val, _ = c.Get("dev_payload")
		pl, _ := val.(*DevReg)
		if c.Request.Method == "POST" {
			// Clear the old logs - get the recent ones, replace the entire array
			// then append the new ones onto the same
			recentLogs := []*DevReg{}
			err := ILogs(pl).QRecentLogs(-1, func(m []bson.M) error {
				return devreg.Pipe(m).All(&recentLogs)
			}) //getting all logs that are a month old
			if err != nil {
				errx.DigestErr(errx.NewErr(&errx.ErrQuery{}, err, "Failed to get recent logs", "HndlDeviceLogs/QRecentLogs"), c)
				return
			}
			if len(recentLogs) > 0 { //when no logs, there is no need for replacing anything
				// Notice how we have changed the base object for query
				// recentLogs is the DevREg with recent logs, do not use payload here
				// payload has the logs that are to be newly pushed
				err = ILogs(&DevReg{SID: pl.SID, LData: recentLogs[0].LData}).QReplaceLogs(func(sel, upd bson.M) error {
					return devreg.Update(sel, upd)
				}) // trimming the entire log array to only a month old logs
				if err != nil {
					errx.DigestErr(errx.NewErr(&errx.ErrQuery{}, err, "Failed to trim logs", "HndlDeviceLogs/QReplaceLogs"), c)
					return
				}
			}
			err = ILogs(pl).QPushLogs(func(sel, upd bson.M) error {
				return devreg.Update(sel, upd)
			})
			if err != nil {
				errx.DigestErr(errx.NewErr(&errx.ErrQuery{}, err, "Failed to push new logs", "HndlDeviceLogs/QPushLogs"), c)
				return
			}
		} else if c.Request.Method == "GET" {
			serial := c.Param("serial")
			qry := c.Query("q")
			/* qry=info/warning/error
			device logs are for a particular device serial filtered for the level
			if no filter then all the logs will just be sent back
			*/
			result := []*DevReg{}
			err := ILogs(&DevReg{SID: serial}).QGetLogs(qry, func(pipe []bson.M) error {
				log.WithFields(log.Fields{
					"result": pipe,
				}).Info("Now logging the query of the GET request")
				return devreg.Pipe(pipe).All(&result)
			})
			if err != nil {
				errx.DigestErr(errx.NewErr(&errx.ErrQuery{}, err, "failed to get filtered logs", "HndlDeviceLogs/QGetLogs"), c)
				return
			}
			if len(result) > 0 {
				c.JSON(http.StatusOK, result[0])
				return
			}
			c.JSON(http.StatusOK, nil) // incase there are no logs
			return
		}
	}
}
