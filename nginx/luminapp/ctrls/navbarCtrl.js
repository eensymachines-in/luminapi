(function(){
    angular.module("luminapp").controller("navbarCtrl", function($scope,lclStorage, $timeout){
        $scope.$on('$routeChangeSuccess', function($event, next, current) {
            if(next.$$route.originalPath =="/"){ //logged out or at the splash page
                // either of cases we dont need the nav links 
                $scope.authInfo = null;
                return
            }
            if (!$scope.authInfo) { //only if the authInfo has never been read
                $scope.authInfo =lclStorage.get_auth() //getting the authinfo 
                if (!$scope.authInfo){ // but the cache may not be set as yet
                    $timeout(function(){ //so trying out after a delay
                        console.log("Now trying out getting local cache after a delay")
                        $scope.$apply(function(){ //since this is outside the controller function 
                            $scope.authInfo =lclStorage.get_auth() // expecting the cache to be set by now
                            if ($scope.authInfo == undefined)  { //if its still not then give up
                                return
                            }
                        })
                    }, 1000) // for the srvapi to push to cache it takes a bit of time
                }
            }
        });
    })
})()