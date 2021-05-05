(function(){
    angular.module("luminapp", ["ngRoute"]).config(function($routeProvider,$interpolateProvider,$provide){
        // with GO Lang frameworks this can help to have angular a distinct space 
        $interpolateProvider.startSymbol("{[")
        $interpolateProvider.endSymbol("]}")
       
        $routeProvider
        .when("/", {
            templateUrl:"/views/splash.html"
        })
        .when("/:email/account", {
            templateUrl:"/views/account.html",
        })
        .when("/:email/devices", {
            templateUrl:"/views/user-devices.html",
        })
        // Schedules of devices
        .when("/:serial/schedules", {
            templateUrl:"/views/device-schedules.html",
        })
        .when("/signup", {
            templateUrl:"/views/signup.html",
        })
        .when("/about", {
            templateUrl:"/views/about.html",
        })
        .when("/admin/:email/devices", {
            templateUrl:"/views/admin-devices.html",
        })
        .when("/admin/accounts", {
            templateUrl:"/views/admin-accs.html",
        }) 
        .when("/admin/embargo", {
            templateUrl:"/views/embargo-devices.html",
        })
        .otherwise({redirectTo:"/"})

        // /^([0-1]\d):([0-5]\d)\s{1}(?:AM|PM)?$/i
        $provide.provider("schedTmPattern", function(){
            this.$get = function(){
                // this pattern can validate the time entered in the schedule table 
                // since we are resorting to manual entry of time the validation is a bit necessary
                return /^([0-1]\d):([0-5]\d)\s{1}(?:AM|PM)?$/i
            }
        })
        // serves up a regex that can help us test and identify a valid email id
        // this will be used by multiple controllers
        $provide.provider("emailPattern", function(){
            this.$get = function(){
                // [\w] is the same as [A-Za-z0-9_-]
                // 3 groups , id, provider , domain also a '.' in between separated by @
                // we are enforcing a valid email id 
                // email id can have .,_,- in it and nothing more 
                return /^[\w-._]+@[\w]+\.[a-z]+$/
            }
        })
        $provide.provider("passwdPattern", function(){
            this.$get = function(){
                // here for the password the special characters that are not allowed are being singled out and denied.
                // apart form this all the characters will be allowed
                // password also has a restriction on the number of characters in there
                return /^[\w-!@#%&?_]{8,16}$/
            }
        })
        $provide.provider("baseURL", function(){
            // change this when the subdomain changes and all the services will follow
            this.$get = function(){
                return {
                    auth : "http://auth.eensymachines.in",
                    // TODO: before moving to production change this uri
                    lumin: "http://localhost/api/v1/devices"
                }
            }
        });
        // TODO: when moving to dev comment this.
        // console.log = function(){};     
        // console.table = function(){}; 

    }).filter("nameFlt", function(){
        return function(name, limit){
            if (name.length> limit){
                return name.slice(0,limit)
            }
            return name
        }
    }).filter("emailFlt", function(){
        return function(name){
           return name.replace(/@[\w]{1,}.com$/, "@..")
        }
    }).filter("locFlt", function(){
        return function(name){
           return name.replace(/[\s]{1,}[\w]{1,}/, "...")
        }
    }).filter("serialFlt", function(){
        return function(serial){
            return serial.replace(/^0+/,'')
        }
    })
})()