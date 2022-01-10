
(function(){
    /*This service exclusively deals with getting an setting authentication in cache*/ 
    angular.module("luminapp").service("lclStorage", function($window){
        this.clear_auth = function(){
            $window.localStorage.removeItem("u-auth")
            $window.localStorage.removeItem("u-refr")
            $window.localStorage.removeItem("u-email")
            $window.localStorage.removeItem("u-role")
            $window.localStorage.removeItem("u-name")
        }
        // this is inline with issue #20: while setting from authorization only the tokens need to be updated in the cache 
        // if the authentication has alreasy set the required creds, authorization need not tamper what is not required 
        // only updating token
        this.set_token_auth = function(auth, refr){
            $window.localStorage.setItem("u-auth", auth)
            $window.localStorage.setItem("u-refr", refr)
        }
        this.set_auth = function (auth, refr, email, role, name){
            $window.localStorage.setItem("u-email", email)
            $window.localStorage.setItem("u-auth", auth)
            $window.localStorage.setItem("u-refr", refr)
            $window.localStorage.setItem("u-role", role)
            $window.localStorage.setItem("u-name", name)
        }
        this.get_auth = function(){
            // if there is no email in the cache, the authentiation returned is undefined
            // undefined authentication signifies 
            return $window.localStorage["u-email"] == undefined ? undefined: {
                email:$window.localStorage["u-email"],
                authtok:$window.localStorage["u-auth"],
                refrtok: $window.localStorage["u-refr"],
                role: $window.localStorage["u-role"],
                name: $window.localStorage["u-name"],
            }
        }
    })
})()