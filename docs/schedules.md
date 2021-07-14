### Schedules CRUD
---------

Modifying schedules involves `Patching schedules` on __existing registered devices__. Such when patched are passed to the MQTT broker and further to subscribing device. Cloud remains the single source of truth, when modified will `push` to synchronize devices (on the ground).

Here we discuss schedules from the standpoint of the web services / application. We start from the database models and then rile up all the way to the front on the user interface. The purpose of this documentation is the detailed view of the code and understanding of the CRUD of schedules on the server

```json
{"_id":{"$oid":"60bccb9841a9aa9a3e859ade"},
"serial":"000000007920365b",
"scheds":[
    {"on":"06:00 PM","off":"09:30 PM","ids":["IN1","IN2","IN3","IN4"],"primary":true},
    {"on":"07:00 PM","off":"06:45 PM","ids":["IN2","IN3"],"primary":false},
    {"on":"06:00 AM","off":"06:30 AM","ids":["IN1"],"primary":false}],
"rmaps":[
    {"rid":"IN1","defn":"Floor1-12-13-8"},
    {"rid":"IN2","defn":"Porch-Floor4-5-7-9"},
    {"rid":"IN3","defn":"Basement"},
    {"rid":"IN4","defn":"Terrace"}],
"logs":[]
}
```
Any device is registered with requisite information 
- schedules 
- relay maps 
- logs

against the unique serial number of the device. This is not the `auth` information but the `reg` information specific to the autolumin application.

#### Schedules :
--------


```
"scheds":[
    {"on":"06:00 PM","off":"09:30 PM","ids":["IN1","IN2","IN3","IN4"],"primary":true},
    {"on":"07:00 PM","off":"06:45 PM","ids":["IN2","IN3"],"primary":false},
    {"on":"06:00 AM","off":"06:30 AM","ids":["IN1"],"primary":false}
]
```

Are just an array of objects each denoting double trigger to the dedvice at a specific time. Plus a schedule has the information on which specific nodes of the relay will the actuation be applied to. A flag marker on the schedule also denotes if the schedule is `primary` or not.

- On time is time as string when the relay will be comanded to be turned `ON`
- Off time is the time as string when the relay will be commanded to be turned `OFF`
- `ids` Collectio of all the relay ids that would be actuated
- Primary schedule / Exceptions have slight operational difference, wrt to current time and the time at which device comes up


#### RMaps :
------

While relays can be addressed by machines using the simple `IDs` those are of no use to the end user. End users would like to see descriptive identification to relays. Infact users have no clue of relays underneath. RMaps are a way of mapping `relay ids` to `user descriptions`. This map is set from the device when self-registering. 

#### Logs :
----------

Device logs are pushed to this array timely so that we save a considerable memory footprint on the ground plus the source of truth is always the cloud. We can later use the log data on the cloud to do a log analysis and bette diagnonsis of the device on the ground.

#### Schedules CRUD:
-----------

Most critical of all the api endpoints in autolumin is the CRUD of schedules of the device. Single source of truth is always the cloud database, device and the user interface just follow it at all times. 

- User requests the device details
- Device details has the schedules
- Schedules are in an array, primary and exceptions 
- User modifies - adds, removes, edits the schedules
- Schedules are patched on the cloud database, and notification pushed onto mosquitto
  - Schedules have no conflict
    - Device listening on the MQTT notification then reads in the new schedule(s)
    - Schedules are re-calculated and service is refreshed for the new schedules.
  - Schedules have conflict
    - HTTP error code sent back with array of coflicting schedules
    - Front end identifies the schedule with conflict 
    - conflicting schedule is prompted to be resolved 
    - new schedule is posted back to the API
  
Understanding how the API is structured will be the first step in getting a firm grasp of the entire concept clearly. 

```go 
devices := r.Group("/devices")
devices.Use(dbConnect())
devices.PATCH("/:serial", checkIfDeviceReg(true), devregPayload, mqttConnect(), HandlDevice)
```
Above is the API endpoint for Patching a schedule of a device 
- gets the device details as payload 
- uses couple of middleware to inject new objects onto the context
- handles the request, makes changes to the device

Below are the signatures of the middleware functions. The names are self evident to tell you what they implement.

```go 
func dbConnect() gin.HandlerFunc {}
func checkIfDeviceReg(abortIfNot bool) gin.HandlerFunc{}
func devregPayload(c *gin.Context) {}
func mqttConnect() gin.HandlerFunc {}
```
Lets then have a closer look at the device handler - PATCH

```
https://github.com/eensymachines-in/luminapi/blob/4b10c4dd88cc1288886724f55dbe413a54fdc643/handlers.go#L79-L157
```
