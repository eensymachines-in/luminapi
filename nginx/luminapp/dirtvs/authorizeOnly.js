(function(){
    angular.module("luminapp").directive("authorizedOnly", function(lclStorage, srvApi, $location, $rootScope){
        return {
            restrict:"A",
            replace:false,
            scope: false,
            transclude:true,
            // the child DOM content is loaded only when the authorized flag is true 
            template:"<div ng-if='authorized==true'><ng-transclude><ng-transclude></div>",
            controller : function($scope){
                $scope.authorized = false; // the child content is delayed in loading
                if (!$scope.lvl) { // this wouldl mean we just want all levels to access this 
                    $scope.lvl =0;
                }
                var upon_err_exit = function(){
                    // anytime the erorr is exited from the err-modal this function will be called back
                    $scope.$apply(function(){
                        $location.url("/")
                    })
                }
                // here we would want to pick the login tokens from the local cache
                $scope.authInfo =lclStorage.get_auth()
                if (!$scope.authInfo){
                    // unless we have complete info there is no point in continuing 
                    $rootScope.err = {
                        message :"Invalid/No authentication details found, kindly login again",
                        status:"",
                        statusText: "Ooops!",
                        upon_exit: upon_err_exit,
                    }
                }else{
                    // this is when we have valid authInfo
                    srvApi.authorize($scope.authInfo.email, $scope.authInfo.authtok, $scope.authInfo.refrtok, $scope.lvl).then(function(data){
                        console.info("Authorized !")
                        $scope.authorized = true; // the child content can now load up
                        return
                    }, function(error){
                        // authorized status need not change since its already set to false
                        error.upon_exit = upon_err_exit;
                        $rootScope.err = error;
                    })
                }
               
            }

        }
    })
})()