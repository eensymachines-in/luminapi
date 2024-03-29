(function(){
    /*schedCtrl:  helps to control a list of schedules [{on:"",off:"",primary:true,ids:["IN1"]}] of a single device
    uuid of the device is from the $routeParams.serial
    A deep watch on the schedules will help to extend the object for the validation functions
    For the first change it also records a comparison JSON string. This when compare to the later state of schedules will let us know if anything has changed*/ 
    angular.module("luminapp").controller("schedCtrl", function($scope, $routeParams,srvApi,srvRefactor,$route,$rootScope){
        $scope.wait = false; //used to show/hide the ribbon progress bar
        $scope.optsSchedules = [];
        $scope.selectedSched = null; // this schedule is the pointer to selected one
        var uuoid = function (){
            // generates a new object id for each new schedule thats added to  $scope.optsSchedules
            // we use this as an id to track the schedules 
            // https://stackoverflow.com/questions/3231459/create-unique-id-with-javascript
            // return Date.now().toString(36) + Math.random().toString(36).substr(2);
            return Math.random().toString(36).substr(2);
        }
        var select_latest_schedule = function(){
            // Not much going on here- only changing the section to latest one
            if($scope.optsSchedules.length>=1){
                $scope.selectedSched = $scope.optsSchedules[$scope.optsSchedules.length-1];
            }else{
                // Select nothing 
                console.error("select_latest_schedule: cannot select if optsSchedules is empty")
                $scope.selectedSched = null;
            }
        }
        $scope.remove_sched = function(id){
            // splice works in-place and returns the item just removed s
            // here all what we do is remove the desired item 
            // id : is the unique id used to identify the objects inside optsSchedules
            $scope.optsSchedules.forEach(function(el,index){
                if (el.oid == id){
                    $scope.optsSchedules.splice(index,1);
                    console.log("Dropping item index: "+ index);
                    return
                }
            })
            select_latest_schedule();
        }
        // deviceDetails.scheds > optsSchedules : object extension
        var extend_api_sched = function(s) {
            // extends the data shape of schedule object from the api to schedule with more derived properties
            // will extend the properties of the sched to enhanced for front end
            // on/off string properties of the schedule are modified from within time-select
            result = angular.extend({}, s)
            result.name = s.primary?"Primary":"Exception-"+$scope.optsSchedules.length;
            result.title =s.primary?"Primary schedule":"Overlay schedule";
            result.desc = s.primary?"Applies to all nodes at once. Exceptions are applied on and over this. While you can change the time here, cannot select the nodes.":"Exception schedules are over and additional to the primary schedule. Select the nodes and the time for which they are applied";
            result.oid= uuoid(); //so that we can track the object quickly when modifying the list
            result.match_time = function(on, off){
                return (this.on == on && this.off == off);;
            }
            result.lbls = function(){
                // getting rmaps definitions from ids that the schedule signifies 
                r = [];
                $scope.deviceDetails.rmaps.forEach(rm => {
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
            }();
            result.has_zero_nodes = function(){
                // this will calculate if the schedule option has no node selected 
                // filters all the labels that have been selected 
                // if length of labels selected ==0 then schedule has zero node selected
                // this will help in identifying this schedule before submit
                // Nodes - IN1, IN2, IN3 , IN3 essentially means nodes on the relay
                return result.lbls.filter(x=>x.sel==true).length ==0 ? true: false
            }
            return result
        }
        // User by clicking on the New button will trigger this
        $scope.new_schedule = function(){
            // Pushes a new default schedule to the list 
            // also selects this latest schedule in the main drop down
            $scope.optsSchedules.push(extend_api_sched({
                on:"01:00 AM",
                off:"01:00 PM",
                ids:[],
                primary:false,
            }));
            select_latest_schedule();
        }
        // srvRefactor($scope).get_object_from_api will trigger this
        $scope.$watch("deviceDetails", function(after, before){
            if (after){
                //  populating the schedTabs array
                after.scheds.forEach((x,i)=>{
                    $scope.optsSchedules.push(extend_api_sched(x))
                }) //this will trigger optsSchedules watch and schedule would be modfied further
                $scope.selectedSched = $scope.optsSchedules[0];
                console.table($scope.optsSchedules);
            }else{
                console.log("deviceDetails: changed but not acknowledged")
                console.log(after);
            }
        }) 
        // GET the device details from the api
        // If it fails to do so, it would result in an error modal
        // Sequence -1 
        srvRefactor($scope).get_object_from_api(function(){
            return srvApi.get_device_schedules($routeParams.serial)
        },function(){
            console.error("Failed to get device schedules");
        }, "deviceDetails") //getting the schedules would get all the device details
        // schedules need the relay maps as well 
        $scope.submit = function(){
            if ($scope.optsSchedules.filter(x=>x.has_zero_nodes()==true).length==0){
                // Normal case when the user has selected atleast one node on each schedule
                scheds = []; // array of schedules ready to be dispatched to the api
                // below we go ahead to transform the optsSchedules model to sched model as required by the api
                $scope.optsSchedules.forEach(function(el){
                    scheds.push({
                        on:el.on, 
                        off:el.off, 
                        primary:el.primary,
                        ids: function(){
                            // this will determine if the user has selected atleast one node
                            // if none are selected we need to mandate selection of atleast one
                            re = [];
                            el.lbls.filter(x=>x.sel ==true).forEach(x=>re.push(x.rid))
                            return re
                        }()
                    })
                });
                // this goes into #16
                payload = {
                    serial: $scope.deviceDetails.serial,
                    scheds: scheds
                }
                srvApi.patch_device_schedules(payload.serial, payload).then(function(data){
                    console.log("Success! schedules have been saved");
                    console.log(data);
                    $route.reload();
                },function(error){
                    if (error.conflicts.length >0){
                        // this case when we have schedules conflicting 
                        error.conflicts.forEach(el=>{
                            pick =$scope.optsSchedules.filter(sch=>sch.match_time(el.on, el.off)==true);
                            if (pick.length>0){
                                pick[0].conflicts = true;
                            }
                        });
                        error.upon_exit = function(){
                            // error dismissed but page not reloaded
                            // this is when 
                            $scope.$apply(function(){})
                        }
                        console.table($scope.optsSchedules);
                    } else{
                        error.upon_exit = function(){
                            // this runs when the modal is dismissed 
                            $scope.$apply(function(){
                                $route.reload()
                            })
                        }
                    }
                   
                    $rootScope.err = error;
                })
            }else {
                // case when you need to enforce selection of atleast one label on every schedule
                $rootScope.err = {
                    statusText: "One or more schedules have no nodes selected",
                    message: "Every schedule needs atleast one area selected on which it operates. Check all the schedules for unselected areas",
                    upon_exit: function(){
                        console.log("Acknowledged error..");
                    }
                }
            }
                
        }
    })
})()