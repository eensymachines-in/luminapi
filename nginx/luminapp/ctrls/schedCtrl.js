(function(){
    /*schedCtrl:  helps to control a list of schedules [{on:"",off:"",primary:true,ids:["IN1"]}] of a single device
    uuid of the device is from the $routeParams.serial
    A deep watch on the schedules will help to extend the object for the validation functions
    For the first change it also records a comparison JSON string. This when compare to the later state of schedules will let us know if anything has changed*/ 
    angular.module("luminapp").controller("schedCtrl", function($scope, $routeParams,srvApi, $route,srvRefactor, $rootScope,schedTmPattern){
        $scope.wait = false; //used to show/hide the ribbon progress bar
        var origSchedules = JSON.stringify({}) //to start with its the imperession of an empty object
        $scope.$watch("schedules", function(after, before){
            if (after!==null && after!==undefined){
                console.log(origSchedules == JSON.stringify({}));
                // also make a json string for later comparison
                if (origSchedules == JSON.stringify({})) {
                    // this can happen only once since we want to compare to the very first schedule set that we get from the api
                    // all the consequent changes in schedules object will not be stringified
                    // to detect any changes made we need to compare the after-change schedules list to first one 
                    // schedules list can change when you add / remove schedules as well.
                    origSchedules = JSON.stringify(after)
                }
                // adding validation functions to the object
                // functions are not stringified so it does not make a difference to the 'change' comparison
                after.forEach(function(el, index){
                    el.validate_on = function(){
                        return schedTmPattern.test(el.on);
                    }
                    el.validate_off = function(){
                        return schedTmPattern.test(el.off);
                    }
                })
            }
        }, true) // here we are attempting to deep watch an array so that when items are added or removed we can detect that 
        
        // GET the schedules list from api 
        // If it fails to do so, it would result in an error modal
        srvRefactor($scope).get_list_from_api(function(){
            return srvApi.get_device_schedules($routeParams.serial)
        }, function(){},function(){
            console.error("Failed to get device schedules");
        }, "schedules") 
        $scope.submit = function(){
            // console.log(JSON.stringify($scope.schedules))
            // console.log(origSchedules)
            if (JSON.stringify($scope.schedules) !== origSchedules) {
                // Here we need to check to see if the original list is necessary before we push the changes.
                // remember that this patch operation would also involve the data broker sending a push notification to the device on the ground
                // console.log("we are about to submit changes to the schedules")
                $scope.wait = true;
                // $scope.schedules.forEach(x=> delete x.conflicts) // incase there has been an conflicts error we dont want the extra properties to go along with the body
                srvApi.patch_device_schedules($routeParams.serial,$scope.schedules).then(function(){
                    console.log("Success.. we have patched the schedules on the cloud");
                    $scope.wait = false;
                    $rootScope.success = {
                        title :"Done!",
                        message:"Schedules updated! - If your device was online, it would have received this change",
                        upon_exit: function(){
                            $route.reload();
                        }
                    }
                },function(error){
                    $scope.wait = false;
                    error.upon_exit = function(){
                        console.error(error);
                        // We wouldnt want the reload since then the changes (although invalid) will not reflect back
                    }
                    if (error.status ==400) {
                        // incase of a bad request we add the conflicts to the respective schedules 
                        console.log(error.conflicts);
                        error.conflicts.forEach(function(el, index){
                            error.message += "\n"+el.on +"-"+el.off;
                        })
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