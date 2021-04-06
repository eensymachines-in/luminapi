package main

import (
	"fmt"
	"net/http"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/eensymachines-in/errx"
	"github.com/eensymachines-in/luminapi/core"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
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
	log.Info("Now connected to MQTT client..")
}
func onMqttConnectLostHandler(client mqtt.Client, err error) {
	log.Warnf("MQTT client connection lost %s", err)
}

// mqttConnect : injects the client into the context - use this when you have to publish
func mqttConnect() gin.HandlerFunc {
	return func(c *gin.Context) {
		opts := mqtt.NewClientOptions()
		opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
		opts.SetClientID("go_mqtt_client")
		if os.Getenv("MQTT_U") == "" || os.Getenv("MQTT_P") == "" {
			// if the password for the the mqtt client was not set the api will result in an error
			errx.DigestErr(errx.NewErr(&errx.ErrEncrypt{}, fmt.Errorf("invalid username or password for the mqtt broker"), "One of our gateways had a error, please try after sometime", "mqttConnect"), c)
			return
		}
		opts.SetUsername(os.Getenv("MQTT_U"))
		opts.SetPassword(os.Getenv("MQTT_P"))
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
