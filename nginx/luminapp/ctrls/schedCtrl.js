(function(){
    angular.module("luminapp").controller("schedCtrl", function($scope, $routeParams,srvApi, $route,srvRefactor, $rootScope,schedTmPattern){
        $scope.wait = false;
        var origSchedules = JSON.stringify({}) //to start with its the imperession of an empty object
        $scope.$watch("schedules", function(after, before){
            if (after!==null && after!==undefined){
                // also make a json string for later comparison
                origSchedules = JSON.stringify(after)
                // adding validation functions to the object
                // functions are not stringified so it does not make a difference to the 'change' comparison
                after.forEach(function(el, index){
                    el.validate = function(val){
                        return schedTmPattern.test(val);
                    }
                })
            }
        })
        // delegate the entire call implementation to boilerplate code
        // handles the implementation at one place
        srvRefactor($scope).get_list_from_api(function(){
            return srvApi.get_device_schedules($routeParams.serial)
        }, function(){},function(){
            console.error("Failed to get device schedules");
        }, "schedules")
        // srvApi.get_device_schedules($routeParams.serial).then(function(data){
        //     console.log("We have received the schedules for the device: " +$routeParams.serial)
        //     console.table(data)
        //     $scope.schedules = data;
        // }) //TODO: implement the error function too .. 
        $scope.submit = function(){
            // console.log(JSON.stringify($scope.schedules))
            // console.log(origSchedules)
            if (JSON.stringify($scope.schedules) !== origSchedules) {
                // Here we need to check to see if the original list is necessary before we push the changes.
                // remember that this patch operation would also involve the data broker sending a push notification to the device on the ground
                // console.log("we are about to submit changes to the schedules")
                $scope.wait = true;
                srvApi.patch_device_schedules($routeParams.serial,$scope.schedules).then(function(){
                    console.log("Success.. we have patched the schedules on the cloud");
                    $route.reload();
                },function(error){
                    error.upon_exit = function(){
                        console.error(error)
                    }
                    $rootScope.err = error;
                })
            }else {
                console.log("Make some changes to the schedules before submit")
            }
        }
        $scope.add_schedule = function(){
            // this shall add a new template schedule to the list
            // has default times on ON and OFF 
            // user can edit them before saving
            $scope.schedules.push({
                "on":"00:00 AM",
                "off":"00:00 PM",
                "ids":["IN1","IN2","IN3","IN4"],
                "primary":false,
            })
        }
        $scope.remove_schedule = function(index){
            // shall remove the schedule from the list 
            // cannot remove primary schedules
            // does no server action
            $scope.schedules =$scope.schedules.filter(function(item,idx){
                return idx!=index;
            })
        }
    })
})()