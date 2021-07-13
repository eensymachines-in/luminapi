### Schedules
---------

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