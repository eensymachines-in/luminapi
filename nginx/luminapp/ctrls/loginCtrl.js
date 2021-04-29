(function(){
    angular.module("luminapp").controller("loginCtrl", function($scope, $rootScope,$route,$location,srvApi,lclStorage,$timeout){
        $scope.wait = false;
        srvApi.log_out().then(function(){
            lclStorage.clear_auth()
        }, function(error){
            lclStorage.clear_auth()
            console.error("There was problem logging out the user: "+error)
        })
        $scope.submit = function(){
            $scope.$broadcast("validate",{})
            $timeout(function(){
                $scope.$apply(function(){
                    if(!$scope.isEmailInvalid && !$scope.isPassInvalid) {
                        $scope.wait = true;
                        srvApi.log_in($scope.details.email,$scope.details.passwd).then(function(data){
                            $scope.err = null;
                            $scope.wait = false;
                            $location.url("/"+$scope.details.email+"/account")
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