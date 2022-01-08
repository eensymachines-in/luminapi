(function(){
    angular.module("luminapp").controller("loginCtrl", function($scope, $rootScope,$route,$location,srvApi,lclStorage,$timeout){
        $scope.wait = false;
        console.log("now logging in the base url: "+$location.host());
        console.log("protocol: "+$location.protocol());
        console.log("abs url: "+$location.absUrl());
        srvApi.log_out().then(function(){
            lclStorage.clear_auth()
        }, function(error){
            lclStorage.clear_auth()
            console.error("There was problem logging out the user: "+error)
        })
        $scope.submit = function(){
            $scope.$broadcast("validate",{});
            $timeout(function(){
                $scope.$apply(function(){
                    // by this time the broadcasted validate command would have been complete 
                    if(!$scope.isEmailInvalid && !$scope.isPassInvalid) {
                        $scope.wait = true;
                        srvApi.log_in($scope.details.email,$scope.details.passwd).then(function(data){
                            $scope.err = null;
                            $scope.wait = false;
                            // Ahead of feedback from the user, it makes more sense to show devices rather than account details
                            $location.url("/"+$scope.details.email+"/devices")
                        }, function(error){
                            error.upon_exit = function(){
                                $scope.$apply(function(){
                                    $route.reload();
                                })
                            }
                            $scope.wait = false;
                            $rootScope.err = error;
                        })
                    }
                })
            },500)
        }
    })
})()