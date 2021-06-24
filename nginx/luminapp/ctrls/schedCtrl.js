(function(){
    /*schedCtrl:  helps to control a list of schedules [{on:"",off:"",primary:true,ids:["IN1"]}] of a single device
    uuid of the device is from the $routeParams.serial
    A deep watch on the schedules will help to extend the object for the validation functions
    For the first change it also records a comparison JSON string. This when compare to the later state of schedules will let us know if anything has changed*/ 
    angular.module("luminapp").controller("schedCtrl", function($scope, $routeParams,srvApi, $route,srvRefactor, $rootScope,schedTmPattern){
        $scope.wait = false; //used to show/hide the ribbon progress bar
        $scope.pSched={on:"",off:""}; // primary schedule
        $scope.$watch("pSched.on", function(after, before){
            if (after && after!==before){
                console.log(after);
            }
        })
        $scope.$watch("pSched.off", function(after, before){
            if (after && after!==before){
                console.log(after);
            }
        })
        unreg=$scope.$watch("deviceDetails", function(after, before){
            if (after!==null && after!==undefined && after !== before){
                after.scheds.forEach(el=>{
                    if (el.primary ==true){
                        $scope.pSched = {on:el.on,off:el.off};
                        return;
                    }
                });
            }
        }) // here we are attempting to deep watch an array so that when items are added or removed we can detect that         
        
        // GET the schedules list from api 
        // If it fails to do so, it would result in an error modal
        srvRefactor($scope).get_object_from_api(function(){
            return srvApi.get_device_schedules($routeParams.serial)
        },function(){
            console.error("Failed to get device schedules");
        }, "deviceDetails") //getting the schedules would get all the device details
        // schedules need the relay maps as well 
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