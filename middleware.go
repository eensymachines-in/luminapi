package main

import (
	"fmt"
	"net/http"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/eensymachines-in/errx"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
		// pack and send in the most important dtabase objects - session.Close() and collection
		c.Set("devreg", session.DB("lumin").C("devreg")) // send in the collection directly
		c.Set("db_close", func() {
			session.Close()
		})
	}
}

// devregPayload : from the incoming requests this will strip the the payload and map it onto the devReg object
func devregPayload(c *gin.Context) {
	payload := &DevReg{}
	if c.ShouldBindJSON(payload) != nil {
		// error unmarshalling json
		errx.DigestErr(errx.NewErr(&errx.ErrJSONBind{}, nil,
			fmt.Sprintf("failed to read the device logs, Expected format : %v", payload), "devLogPayload/ShouldBindJSON"), c)
		return
	}
	log.WithFields(log.Fields{
		"payload": payload,
	}).Info("Received device registration payload")
	c.Set("dev_payload", payload) // sedn it packing to the next handler
}

// checkIfDeviceReg : checks to see if the device is registered, depending on how its configured it will either abort or continue with injecting into the context
// abortIfNot : set to true and the device is not registerd the handler will abort else will continue
// ** Please note this middleware has to be preceeded with  lclDbConnect -since the device needs to be checked against a database
// Incase of POST request  this has also to be preceeded by devregPayload - since the serial number of the device is in the payload
func checkIfDeviceReg(abortIfNot bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("devreg")
		if !exists {
			errx.DigestErr(errx.NewErr(&errx.ErrConnFailed{}, nil, "failed connection to database", "HandlDevices/dbConnect"), c)
			return
		}
		devreg, _ := val.(*mgo.Collection)
		pl := &DevReg{}
		if c.Request.Method != "POST" {
			// Only post requests do not have the serial as a param
			pl.SID = c.Param("serial")
		} else {
			// Within the context of post request the serial number is in the payload and not in the param
			val, _ = c.Get("dev_payload") // has to be preceeded with devregPayload else will not get this payload
			pl, _ = val.(*DevReg)
		}
		var yes bool
		err := pl.OfSerial(func(flt bson.M) error {
			count, _ := devreg.Find(flt).Count()
			yes = count > 0
			return nil
		})
		if errx.DigestErr(err, c) != 0 {
			return
		}
		if !yes && abortIfNot {
			// the device you are looking for is not registered
			// also the calling handler needs this to abort if not registered
			errx.DigestErr(errx.NewErr(&errx.ErrNotFound{},
				fmt.Errorf("no device with serial %s", pl.SID),
				"Device you are looking for is not registered",
				"HandlDevice/IsRegistered"), c)
			// Since no further handlers would execute, this shall close the database connection
			val, _ = c.Get("db_close")
			val.(func())()
			return
		}
		// We wouldnt want to close the database connection in this case
		// there are furtther handlers in line
		// case when whatever the registration state the handler is not to abort
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
		/*Check the init function of the application, here in the middleware we are expecting the username and the password to be loaded in the container environment*/
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
		// check the handler for disconnect.
		// It is important to disconnect and dispose the client
		// unlike the database connection, this would be disposed in the handler
		c.Set("mqttclient", client)
	}

}
