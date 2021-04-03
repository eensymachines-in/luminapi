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

### Device and Webapp as clients:
--------

While the api microservice is central to both the webapp and the clients, its design needs to be open for extension, closed for modification. Webapp and the device are just end clients that either get the data or set the data. These 2 items in no way shall enforce design constraints and data model.

### Api microservice with the core package
--------

The API microservice shall have the core package that would be a shared package containing the business logic of the application. Device / webapp may use it for referencing data models or service functions.


Now the question remains, from the device perspective should this api, communicate further to the authapi and register the device ? or or should it be the device's responsibility to do the same independently ?

#### Encapsulated `authapi` 
------
This means any changes to the authapi shall enforce changes to __luminapi only__, and leave the devices unaffected. Since now the device has dependency on only the `luminapi` and not on the `authapi` this reduces the impact of changes. Luminapi in this case shall have to provide all the encapsulating functions and act like a call fwding microservice. This in my opinion is a serious duplication of code and effort. And given the microservices approach its grossly unnecessary.

#### Peerside `authapi` 
------

This denotes `luminapi` is agnostic of `authapi` and its the device/app that independently verifies the registeration with 2 different microservices. While in this method we can avoid the code duplication, any changes on `authapi` can and will affect the the clients (webapp and the device). While its easy to reconfing the webapp, pushing changes to the device could be really intricate. 

#### Recommendation :
-------

Peerside design is fine, authapi can be versioned and device code need not change at all. 


### Registering a new device:
--------
Besides registering itself on the authapi server, the device will also attempt to register itself on this microservice as well. When doing so it receives a set of default schedules to start with. Default schedules are generated on the server side and sent back to the device. _This is to maintain the server side as the single source of truth (SSoT)


 



