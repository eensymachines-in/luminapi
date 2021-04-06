package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/eensymachines-in/errx"
	"github.com/eensymachines-in/luminapi/core"
	"github.com/eensymachines-in/scheduling"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

const (
	mgoIP     = "srvmongo"
	dbTimeout = 10 * time.Second // this is when we want to timeout the database connection tru
	broker    = "mosquitto"
	port      = 1883
	mqtt_u    = "eensy"
	mqtt_p    = "10645641993"
)

// CORS : this allows all cross origin requests
func CORS(c *gin.Context) {
	// First, we add the headers with need to enable CORS
	// Make sure to adjust these headers to your needs
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-Type", "application/json")
	// Second, we handle the OPTIONS problem
	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		// Everytime we receive an OPTIONS request,
		// we just return an HTTP 200 Status Code
		// Like this, Angular can now do the real
		// request using any other method than OPTIONS
		c.AbortWithStatus(http.StatusOK)
	}
}
func dbConnect() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := mgo.DialWithInfo(&mgo.DialInfo{
			Addrs:   []string{mgoIP},
			Timeout: dbTimeout,
		})
		if err != nil {
			errx.DigestErr(err, c)
			return
		}
		log.WithFields(log.Fields{
			"address": "localhost:37017",
		}).Info("We now have a connection to the database")
		c.Set("devreg", &core.DevRegsColl{Collection: session.DB("lumin").C("devreg")})
		c.Set("db_close", func() {
			session.Close()
		})
	}
}

// devregPayload : this binds the incoming dev reg payload and injects the same to the context
func devregPayload() gin.HandlerFunc {
	return func(c *gin.Context) {
		payload := &core.DevRegHttpPayload{}
		if c.ShouldBindJSON(payload) != nil {
			errx.DigestErr(errx.NewErr(&errx.ErrJSONBind{}, nil, "failed to read the device registration", "devregPayload/ShouldBindJSON"), c)
			return
		}
		log.WithFields(log.Fields{
			"payload": payload,
		}).Info("Received payload")
		c.Set("devregpayload", payload)
	}
}

// checkIfDeviceReg : checks to see if the device is registered, depending on how its configured it will either abort or continue with injecting into the context
// abortIfNot : set to true and the device is not registerd the handler will abort else will continue
// ** Please note this middleware has to be preceeded with  lclDbConnect -since the device needs to be checked against a database
// Incase of POST request  this has also to be preceeded by devregPayload - since the serial number of the device is in the payload
func checkIfDeviceReg(abortIfNot bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("devreg")
		if !exists {
			errx.DigestErr(errx.NewErr(errx.ErrConnFailed{}, nil, "failed connection to database", "HandlDevices/dbConnect"), c)
			return
		}
		devreg, _ := val.(*core.DevRegsColl)
		var serial string
		if c.Request.Method != "POST" {
			// Only post requests do not have the serial as a param
			serial = c.Param("serial")
		} else {
			// Within the context of post request the serial number is in the payload and not in the param
			val, _ = c.Get("devregpayload") // has to be preceeded with devregPayload else will not get this payload
			payload, _ := val.(*core.DevRegHttpPayload)
			serial = payload.Serial
		}
		yes, err := devreg.IsRegistered(serial)
		if errx.DigestErr(err, c) != 0 {
			return
		}
		if !yes && abortIfNot {
			// the device you are looking for is not registered
			// also the calling handler needs this to abort if not registered
			errx.DigestErr(errx.NewErr(&errx.ErrNotFound{},
				fmt.Errorf("no device with serial %s", serial),
				"Device you are looking for is not registered",
				"HandlDevice/IsRegistered"), c)
			// Since no further handlers would execute, this shall close the database connection
			val, _ = c.Get("db_close")
			val.(func())()
			return
		}
		// We wouldnt want to close the database connection in this case
		// there are furtther handlers in line
		c.Set("isreg", yes)
	}
}
func onMqttConnectHandler(client mqtt.Client) {
	fmt.Println("Connected")
}
func onMqttConnectLostHandler(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

// mqttConnect : injects the client into the context - use this when you have to publish
func mqttConnect() gin.HandlerFunc {
	return func(c *gin.Context) {
		opts := mqtt.NewClientOptions()
		opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
		opts.SetClientID("go_mqtt_client")
		opts.SetUsername(mqtt_u)
		opts.SetPassword(mqtt_p)
		// opts.SetDefaultPublishHandler(messagePubHandler)
		opts.OnConnect = onMqttConnectHandler
		opts.OnConnectionLost = onMqttConnectLostHandler
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			// failed to connect to the mqtt client
			errx.DigestErr(token.Error(), c)
			return
		}
		c.Set("mqttclient", client)
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
	}
}
func main() {

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(CORS)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hi from inside luminapi",
		})
	})
	devices := r.Group("/devices")
	devices.Use(dbConnect())
	devices.POST("/", devregPayload(), checkIfDeviceReg(false), HandlDevices)     // to register new devices
	devices.DELETE("/:serial", checkIfDeviceReg(true), HandlDevice)               // single device un-register
	devices.PATCH("/:serial", checkIfDeviceReg(true), mqttConnect(), HandlDevice) // schedules are updated here
	log.Info("Starting luminapi service ..")
	log.Fatal(r.Run(":8080"))
}
