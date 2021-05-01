
(function(){
    angular.module("luminapp").service("srvApi",function($timeout, $q, baseURL, $http, $window,lclStorage){
        /*Use this $http.then error handler so that we can break down the error response (specific to the server into error response that is used by the web app)*/ 
        var err_message = function(response){
            // err_message : breaks down the error response as required for modals / warning
            var m = "Server unreachable, or responded invalid. Kindly wait for admins to fix this";
            var l = "";
            if (response.data !== null && response.data!==undefined) {
                if (response.data.message !==null && response.data.message!==undefined) {
                    m = response.data.message.split("\n")[0];
                    l=response.data.message.split("\n")[1];
                }
            }
            return {
                "status": response.status,
                "statusText": response.statusText,
                "message": m,
                "logid": l
            }
        }
        var execute_request = async function (req, defered){
            // This is only execute the $http request and resolve / reject depending on the http response
            $http(req).then(function(response){
                defered.resolve(response.data);
            },function(response){
                defered.reject(err_message(response))
            })
        }
        this.get_device_schedules = function(serial){
            var defered  = $q.defer();
            execute_request({
                method :"GET",
                url:baseURL.lumin+"/"+serial,
                headers:{
                    'Content-Type': "application/json",
                },
            }, defered)
            return defered.promise;
        }
        this.patch_device_schedules = function(serial, schedules){
            var defered  = $q.defer();
            execute_request({
                method :"PATCH",
                url:baseURL.lumin+"/"+serial,
                headers:{
                    'Content-Type': "application/json",
                },
                data:JSON.stringify(schedules)
            }, defered)
            return defered.promise;
        }
        // Queries the black list collection and gets all the devices from blacklist 
        this.get_device_blacklist = function(){
            var defered  = $q.defer();
            execute_request({
                method :"GET",
                url:baseURL.auth+"/devices?black=true",
                headers:{
                    'Content-Type': "application/json",
                },
            }, defered)
            return defered.promise;
        }
        this.blacklist_device = function(serial, black){
             // Patches the device for lock / unlock status 
             var defered  = $q.defer();
             authInfo = lclStorage.get_auth()
             execute_request({
                method :"PATCH",
                url:baseURL.auth+"/devices/"+serial+"?black="+black,
                headers:{
                    'Content-Type': "application/json",
                    'Authorization': "Bearer "+ authInfo.authtok,
                }
            },defered)
            return defered.promise;
        }
        this.lock_device = function (serial, lock){
            // Patches the device for lock / unlock status 
            var defered  = $q.defer();
            authInfo = lclStorage.get_auth()
            execute_request({
                method :"PATCH",
                url:baseURL.auth+"/devices/"+serial+"?lock="+lock,
                headers:{
                    'Content-Type': "application/json",
                    'Authorization': "Bearer "+ authInfo.authtok,
                }
            }, defered)
            return defered.promise;
        }
        this.get_user_devices = function(email) {
            // For any given user this shall get the devices there in
            var defered  = $q.defer();
            execute_request({
                method :"GET",
                url:baseURL.auth+"/users/"+email+"/devices",
                headers:{
                    'Content-Type': "application/json",
                },
            },defered)
            return defered.promise;
        }
        this.post_acc = function(details) {
            var defered  = $q.defer();
            execute_request({
                method :"POST",
                url:baseURL.auth+"/users",
                headers:{
                    'Content-Type': "application/json",
                },
                data:JSON.stringify(details)
            },defered)
            return defered.promise;
        }
        // This will send a request to 
        this.patch_acc = function(email, passwd){
            var defered  = $q.defer();
            var b64Encoded = btoa(email+":"+passwd)
            execute_request({
                method :"PATCH",
                url:baseURL.auth+"/users/"+email,
                headers:{
                    'Content-Type': "application/json",
                    'Authorization': "Basic "+ b64Encoded
                }
            },defered)
            return defered.promise;
        }
        this.put_acc = function(newAccDetails){
            var defered  = $q.defer();
            authInfo = lclStorage.get_auth()
            execute_request({
                method :"PUT",
                url:baseURL.auth+"/users/"+newAccDetails.email,
                headers:{
                    'Content-Type': "application/json",
                    'Authorization': "Bearer "+ authInfo.authtok
                },
                data:JSON.stringify(newAccDetails)
            },defered)
            return defered.promise;
        }
        this.remove_acc = function(e){
            var defered  = $q.defer();
            var lclAuth = lclStorage.get_auth()
            if (!lclAuth.email) {
                defered.reject({
                    status:403,
                    statusText:"Unauthrized",
                    message:"User not found signed in, please sign in to continue",
                    logid:00,
                })
                return
            }else {
                execute_request({
                    method :"DELETE",
                    url:baseURL.auth+"/users/"+e,
                    headers:{
                        'Content-Type': "application/json",
                        'Authorization': "Bearer "+ lclAuth.authtok
                    }
                },defered)
            }
            return defered.promise;
        }
        this.list_accs = function(){
            // Will get the list of all the account details from the server
            // but needs lvl=2 authorization to do the same
            var defered  = $q.defer();
            var lclAuth = lclStorage.get_auth()
            if (!lclAuth.email) {
                defered.reject({
                    status:403,
                    statusText:"Unauthrized",
                    message:"User not found signed in, please sign in to continue",
                    logid:00,
                })
                return
            }else {
                execute_request({
                    method :"GET",
                    url:baseURL.auth+"/users",
                    headers:{
                        'Content-Type': "application/json",
                        'Authorization': "Bearer "+ lclAuth.authtok
                    }
                }, defered)
            }
            return defered.promise;
        }
        this.get_acc = function(email){
            var defered  = $q.defer();
            execute_request({
                method :"GET",
                url:baseURL.auth+"/users/"+email,
                headers:{
                    'Content-Type': "application/json",
                }
            }, defered)
            return defered.promise;
        }
        this.authorize = function(email, auth, refr, lvl){
            if (auth == refr) {
                console.error("Tokens are identical, this cannot bes")
            }
            var defered  = $q.defer();
            var request  = {
                method :"GET",
                url:baseURL.auth+"/authorize?lvl="+lvl,
                headers:{
                    'Content-Type': "application/json",
                    'Authorization': "Bearer "+ auth
                }
            }
            $http(request).then(function(response){
                defered.resolve({}) // authorized , status is 200 ok
            }, function(response){
                if (response.status == 401) {
                    // the token has expired, now proceeding to refresh the tokens
                    request.url = baseURL.auth+"/authorize?refresh=true"
                    request.headers = {
                        'Content-Type': "application/json",
                        'Authorization': "Bearer "+ refr
                    }
                    $http(request).then(function(response){
                        lclStorage.set_token_auth(response.data.auth, response.data.refr)
                        defered.resolve({})
                    }, function(error){
                        // failed to refresh the authorization
                        // here we dont care whether its 401 or not, - in any
                        defered.reject(err_message(response))
                    })
                }else {
                    // incase there was an error apart from 401 
                    defered.reject(err_message(response))
                }
            })
            return defered.promise;
        }
        this.log_out = function(){
            var defered  = $q.defer();
            authInfo = lclStorage.get_auth()
            if (authInfo !==undefined){
                var request  = {
                    method :"DELETE",
                    url:baseURL.auth+"/authorize",
                    headers:{
                        'Content-Type': "application/json",
                        'Authorization': "Bearer "+ authInfo.authtok
                    }
                }
                // first we logout the authentication token
                $http(request).then(function(response){
                    request.url += "?refresh=true"
                    request.headers  = {
                        'Content-Type': "application/json",
                        'Authorization': "Bearer "+ authInfo.refrtok
                    }
                    $http(request).then(function(){
                        defered.resolve({}) // refresh token also logged out
                    }, function(response){
                        // error logging out the refresh token
                        defered.reject(err_message(response))    
                    })
                }, function(response){
                    // error logging out the auth token
                    defered.reject(err_message(response))
                })
            }else{
                defered.resolve({})
            }
            return defered.promise;
        }
        this.log_in = function(email, passwd){
            var defered  = $q.defer();
            // https://stackoverflow.com/questions/41431429/how-to-decode-base64-encoded-data-into-ascii-in-angularjs
            var b64Encoded = btoa(email+":"+passwd) // since when logging in you need to base64 encode the email and pass
            var request  = {
                method :"POST",
                url:baseURL.auth+"/authenticate/"+email,
                headers:{
                    'Content-Type': "application/json",
                    'Authorization': "Basic "+ b64Encoded
                }
            }
            $http(request).then(function(response){
                // send this to the browser localdb from where it can be picked up 
                // console.log(response.data);
                lclStorage.set_auth(response.data.auth, response.data.refr, response.data.email,response.data.role, response.data.name)
                defered.resolve(response.data);
            },function(response){
                console.log(response)
                defered.reject(err_message(response))
            })
            return defered.promise;
        }
        // This is to test the on fail response of controllers .. 
        // makes no request to the actual api, just fakes a failure response in 1.7 seconds
        // to be used in testing
        this.mock_api_fail = function(mockStatus){
            var defered  = $q.defer();
            $timeout(function(){
                defered.reject({
                    "status": mockStatus,
                    "statusText": "Mock Error",
                    "message": "Mock error from testing! - switch your calls back to calling actula api",
                    "logid": "gfdgdg5655465"
                })
            }, 1700)
            return defered.promise;
        }
    })
})()