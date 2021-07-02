(function(){
    /*schedCtrl:  helps to control a list of schedules [{on:"",off:"",primary:true,ids:["IN1"]}] of a single device
    uuid of the device is from the $routeParams.serial
    A deep watch on the schedules will help to extend the object for the validation functions
    For the first change it also records a comparison JSON string. This when compare to the later state of schedules will let us know if anything has changed*/ 
    angular.module("luminapp").controller("schedCtrl", function($scope, $routeParams,srvApi, $route,srvRefactor, $rootScope,$timeout){
        $scope.wait = false; //used to show/hide the ribbon progress bar
        $scope.optsSchedules = [];
        $scope.selectedSched = null; // this schedule is the pointer to selected one
        $scope.$watch("selectedSched.on", function(after){
            console.log("Change in selected schedule");
            console.log(after)
        })
        $scope.$watch("selectedSched.off", function(after){
            console.log("Change in selected schedule");
            console.log(after)
        })
        var select_latest_schedule = function(){
            $scope.selectedSched = $scope.optsSchedules[$scope.optsSchedules.length-1];
        }
        var extend_api_sched = function(s,i) {
            // extends the data shape of schedule object from the api to schedule with more derived properties
            // will extend the properties of the sched to enhanced for front end
            result = angular.extend({}, s)
            result.name = s.primary?"primary":"schedule";
            result.title =s.primary?"Primary schedule":"Overlay schedule";
            result.desc = s.primary?"Is a wide policy, applied onto all the nodes. Apply individual node exceptions ahead of this. Cannot delete but only modify the primary schedule.":"This policy is applied atop the primary schedule. Its an exception for the specific nodes. Can be deleted and modified.";
            result.remove = s.primary? function(){} : function(){remove_sched(i)};
            result.lbls = function(){
                // getting rmaps definitions from ids that the schedule signifies 
                r = [];
                $scope.deviceDetails.rmaps.forEach(rm => {
                    console.log("Now evaluating the value for s:")
                    console.log(s)
                    fltIds =s.ids.filter(el=>rm.rid ==el);
                    if (fltIds.length >0) {
                        // Relay id is applicable to the schedule
                        item = {txt:rm.defn, sel:true, rid:rm.rid}
                    }else{
                        // relay id is not applicable to the schedule
                        item = {txt:rm.defn, sel:false, rid: rm.rid}
                    }
                    if (s.primary == false){
                        item.togg = function(){
                            this.sel = !this.sel;
                        }
                    }
                    r.push(item)
                });
                return r
            }()
            return result
        }
        var remove_sched = function(schedIndex){
            // splice works in-place and returns the item just removed s
            // here all what we do is remove the desired item 
            console.log("Now removing schedule number :"+ schedIndex);
            $scope.optsSchedules.splice(schedIndex,1);
            select_latest_schedule();
        }
        $scope.new_schedule = function(){
            $scope.optsSchedules.push(extend_api_sched({
                on:"01:00 AM",
                off:"01:00 PM",
                ids:[],
                primary:false,
            },$scope.optsSchedules.length));
            select_latest_schedule();
        }
        $scope.$watch("deviceDetails", function(after, before){
            if (after){
                //  populating the schedTabs array
                after.scheds.forEach((x,i)=>{
                    $scope.optsSchedules.push(extend_api_sched(x,i))
                }) //this will trigger optsSchedules watch and schedule would be modfied further
                $scope.selectedSched = $scope.optsSchedules[0];
                console.table($scope.optsSchedules);
            }else{
                console.log("deviceDetails: changed but not acknowledged")
                console.log(after);
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
    })
})()