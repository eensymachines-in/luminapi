(function(){
    angular.module("luminapp").controller("postAccCtrl", function($scope,srvApi,$route,$timeout,$rootScope, $location){
        $scope.wait = false;
        $rootScope.err = null;
        $scope.submit = function(){
            $scope.wait = true;
            $scope.$broadcast("validate", {})
            $timeout(function(){ // some timeout till the child directives finish updating the invalid status
                $scope.$apply(function(){
                    if($scope.isEmailInvalid || $scope.isPassInvalid){
                        $scope.wait = false;
                        return 
                    } // this is updated in the lower directives 
                    srvApi.post_acc($scope.details).then(function(data){
                        $scope.wait = false;
                        $location.url("/");
                    }, function(error){
                        $scope.wait = false;
                        console.error("Failed to post new account "+ error)
                        error.upon_exit  = function(){
                            $scope.$apply(function(){
                                $route.reload()
                            })
                        }
                        $rootScope.err = error;
                    })
                })
            }, 500)
        }
    })
    .controller("patchAccCtrl", function($scope, $rootScope,lclStorage,srvApi,$timeout,$route, $location){
        // Patches the account for the password
        $scope.wait = false;
        $rootScope.err = null;
        var email = lclStorage.get_auth().email;
        $scope.submit = function(){
            $scope.wait = true;
            $scope.$broadcast("validate", {})
            $timeout(function(){ // some timeout till the child directives finish updating the invalid status
                $scope.$apply(function(){
                    if($scope.isPassInvalid || email == "" || email ==undefined){
                       return 
                    } // this is updated in the lower directives 
                    srvApi.patch_acc(email, $scope.details.passwd).then(function(data){
                        console.log("Account patched..")
                        $scope.wait = false;
                        $location.path("/"); // password changed the authentication is dirty hence login again
                    }, function(error){
                        console.error("Failed to patch account "+ error);
                        $scope.wait = false;
                        error.upon_exit  = function(){
                            $scope.$apply(function(){
                                $route.reload()
                            })
                        }
                        $rootScope.err = error;
                    })
                })
            }, 500)
        }
    })
    .controller("putAccCtrl", function($scope,srvApi,lclStorage,$route,$rootScope,$timeout){
        $scope.wait = false;
        // we then get the email from the local auth 
        userEmail =lclStorage.get_auth().email
        if (userEmail !=undefined && userEmail != "") {
            // then proceed to get the account details
            srvApi.get_acc(userEmail).then(function(data){
                $scope.details = data
            }, function(error){
                upon_error(error)
            })
        }
        $scope.submit = function(){
            if( $scope.details!==undefined && $scope.details !==null && $scope.details!=={}){
                // Only if the details are well defined 
                $scope.wait = true;
                $scope.$broadcast("validate", {}) // child directives can perform the validity check
                // give a fraction of second to all the child directives to set the invalidation data
                $timeout(function(){
                    $scope.$apply(function(){
                        srvApi.put_acc($scope.details).then(function(data){
                            $scope.wait = false
                            $route.reload();
                        }, function(error){
                            $scope.wait = false;
                            error.upon_exit = function(){
                                // this runs when the modal is dismissed 
                                $scope.$apply(function(){
                                    $route.reload()
                                })
                            }
                            $rootScope.err = error;
                        })
                    })
                }, 500)
            }            
        }
    })
})()