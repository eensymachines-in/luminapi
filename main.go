package main

import (
	"bufio"
	"flag"
	"os"
	"strings"
	"time"

	utl "github.com/eensymachines-in/utilities"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	mgoIP       = "srvmongo"
	dbTimeout   = 10 * time.Second // this is when we want to timeout the database connection tru
	broker      = "mosquitto"
	port        = 1883
	fMqttsecret = "/run/secrets/mqtt_secret"
)

var (
	// Flog : determines if the direction of log output
	Flog bool
	// FVerbose :  determines the level of log
	FVerbose bool
)

// Setting the environment variables here and prepare before the api server runs
func init() {
	/*Reading in the command line params, at the entrypoint*/
	utl.SetUpLog()
	flag.BoolVar(&Flog, "flog", true, "direction of log messages, set false for terminal logging. Default is true")
	flag.BoolVar(&FVerbose, "verbose", false, "Determines what level of log messages are to be output, Default is info level")

	file, err := os.Open(fMqttsecret)
	if err != nil {
		log.Warnf("Failed to read encryption secrets, please load those %s", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	// ++++++++++++++++++++++++++++++++ reading in the mqtt secret
	line, _, err := reader.ReadLine()
	if err != nil {
		log.Warn("Error reading the auth secret from file")
	}
	// split the line to read the username and password
	if !strings.Contains(string(line), ":") {
		// Incase the secret file is not as expected
		log.Warn("mqtt_secret: file not in expected format, expected format username:password, without any white space")
	}
	result := strings.Split(string(line), ":")
	if len(result) == 2 {
		log.WithFields(log.Fields{
			"mqtt_u": result[0],
			"mqtt_p": result[1],
		}).Info("we have now loaded the mqtt credentials")
		// Setting the environment variables
		os.Setenv("MQTT_U", result[0])
		os.Setenv("MQTT_P", result[1])
	}
}
func main() {
	/*This will setup the logging direction and depth*/
	flag.Parse()
	logFile := os.Getenv("LOGF")
	if logFile != "" {
		closeLogFile := utl.CustomLog(Flog, FVerbose, logFile) // Log direction and the level of logging
		defer closeLogFile()
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(CORS)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hi from inside luminapi",
		})
	})
	// I would like to keep the url open for future expansion
	// incase we woudl want to launch a newer version of the api whilst keeping the older version this can be our window
	// this versioning can be handled from the nginx server
	// api := r.Group("/api")
	// v1 := api.Group("/v1")
	// ++++++++++++ devices
	devices := r.Group("/devices")
	devices.Use(dbConnect())
	devices.POST("/", devregPayload(), checkIfDeviceReg(false), HandlDevices)     // to register new devices
	devices.DELETE("/:serial", checkIfDeviceReg(true), HandlDevice)               // single device un-register
	devices.PATCH("/:serial", checkIfDeviceReg(true), mqttConnect(), HandlDevice) // schedules are updated here
	devices.GET("/:serial", HandlDevice)                                          // GETting the schedules for a device

	log.Info("Starting luminapi service ..")
	defer log.Warn("Now quitting luminapi service")
	log.Fatal(r.Run(":8080"))
}
