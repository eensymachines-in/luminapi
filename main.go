package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
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
	pages       = "/var/www/luminapp/pages" // this has to be loaded into the container from environment
)

var (
	// Flog : determines if the direction of log output
	Flog bool
	// FVerbose :  determines the level of log
	FVerbose bool
)

// sendIndexHtml : handler for all the request to send back the index.html page
// you can customize this later to have the og:tags modified for SEO and preview links
func sendIndexHtml(c *gin.Context) {
	log.Println("We have reached sendIndexHtml")
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

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
			"mqtt_p": "*******", //result[1],
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
		file, err := os.Open(os.Getenv("LOGF"))
		if err != nil {
			log.Fatal(err)
		}
		gin.DisableConsoleColor()
		gin.DefaultWriter = io.MultiWriter(file)
		defer file.Close()
		defer closeLogFile()
	}
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// server needs access to the index.html and other error pages
	// cause on all the routes the main page is to be sent from the server
	// SEO and image og meta tags will be modified from the server side
	r.LoadHTMLGlob(fmt.Sprintf("%s/*", pages))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"app":     "luminapi",
			"logs":    logFile,
			"verblog": FVerbose,
		})
	})
	r.GET("/", sendIndexHtml)
	r.GET("/signup", sendIndexHtml)
	r.GET("/about", sendIndexHtml)
	r.GET("/admin/accounts", func(c *gin.Context) {
		log.Println("Inside accounts list administration..")
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	r.GET("/admin/embargo", sendIndexHtml)
	r.GET("/admin/devices", sendIndexHtml)
	r.GET("/accounts/:email", func(c *gin.Context) {
		log.Printf("Email param: %s", c.Param("email"))
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	r.GET("/accounts/:email/devices", sendIndexHtml)
	r.GET("/schedules/:serial", sendIndexHtml)

	r.Use(CORS)
	// r.GET("/:email/devices", sendIndexHtml)
	/*Admin related tasks here under one group. Check the nginx conf this has been appropriately */
	grpAdmin := r.Group("/admin")
	grpAdmin.GET("/logs", HndlLogs(os.Getenv("LOGF")))
	// +++++++++++++++++++++ logs
	// Only device specific logs, GET and POST
	logs := r.Group("/logs")
	logs.Use(dbConnect())
	logs.GET("/:serial", checkIfDeviceReg(true), HndlDeviceLogs())
	logs.POST("", devregPayload, checkIfDeviceReg(true), HndlDeviceLogs())
	// ++++++++++++ devices
	devices := r.Group("/devices")
	devices.Use(dbConnect())
	devices.POST("", devregPayload, checkIfDeviceReg(false), HandlDevices)                       // to register new devices
	devices.DELETE("/:serial", checkIfDeviceReg(true), HandlDevice)                              // single device un-register
	devices.PATCH("/:serial", checkIfDeviceReg(true), devregPayload, mqttConnect(), HandlDevice) // schedules are updated here
	devices.GET("/:serial", checkIfDeviceReg(true), HandlDevice)                                 // GETting the schedules for a device

	// a group to facilitate commands to the device from the app
	// A command can only be only posted, the nature of the action in
	cmds := r.Group("/cmds")
	cmds.POST("/:serial", mqttConnect(), HndlCommands)

	log.Info("Starting luminapi service ..")
	defer log.Warn("Now quitting luminapi service")
	log.Fatal(r.Run(":8080"))
}
