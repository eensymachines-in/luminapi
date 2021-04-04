package main

import (
	"net/http"
	"time"

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
		payload := &core.DevRegHttpPayload{}
		if c.ShouldBindJSON(payload) != nil {
			errx.DigestErr(errx.NewErr(&errx.ErrJSONBind{}, nil, "failed to read the device registration", "HandlDevices/ShouldBindJSON"), c)
			return
		}
		log.WithFields(log.Fields{
			"payload": payload,
		}).Info("Received payload")
		yes, err := devreg.IsRegistered(payload.Serial)
		if errx.DigestErr(err, c) != 0 {
			return
		}
		if yes {
			// this indicates the device is already registered
			// the payload is empty
			c.JSON(http.StatusOK, []*scheduling.JSONRelayState{})
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
	devices.POST("/", HandlDevices) // to register new devices
	log.Info("Starting luminapi service ..")
	log.Fatal(r.Run(":8080"))
}
