### Luminapi purpose
-------
Ahead of device and user authentication, we now need a microservice api that can handle CRUD schedules for autolumin. Luminapi is to be accessed by all devices running autolumin. authapi was for any device but this has specific purpose for maitaining the schedules of each of the devices.

1. Refer a database of schedules, filtered by the device ID and send it as json 
2. Modify a schedule in the database, filtered by the device ID and the notify the broker (`mosquitto`) of the same.
3. Clean all schedules for a a device ID
4. When registering a new device a default schedule is injected and responded back to as http response 
5. Endpoint for the devices to communication the status / logs to the cloud.

> Database in the cloud with devices and their schedules is the single source of truth 

Devices will try to stay sync to changes on the cloud via a broker. Changes the user makes shall be pushed on the server and further onto the devices via the broker. The logs and the readings that the device records on the ground can be sent over to the cloud for user reporting.

### Schedule structure :
----------

For a single device this is what the schedule looks like
 - One and only one `primary schedule` is required
 - 
```
{
    "schedules": [
        {"on":"06:00 PM", "off":"07:15 AM","primary":true, "ids":["IN1","IN2","IN3","IN4"]},
        {"on":"04:30 AM", "off":"10:00 PM","primary":false, "ids":["IN4"]}
    ]
}
```