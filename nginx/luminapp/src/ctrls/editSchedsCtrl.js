/*11-FEB-2022 kneerunjun
directives that work to help edit the set of schedules on a device 
refer to view "/schedules/:serial" device-schedules.html
This lets edit all the schedules on a device, add more exceptions, edit primary ones and send back the modification to the api
2 directives - one to manage device details, while other to select edit schedules*/
(function() {
    angular.module("luminapp").directive("deviceDetails", function() {
        /*Responsible to get the device details and provide a dd object on the scope*/
        return {
            restrict: "A",
            scope: false,
            controller: function($scope, srvRefactor, $routeParams, srvApi) {
                $scope.$watch("dd", function(after, before) {
                        if (after !== null && after !== undefined) {
                            /*submit allows the device details to be submitted to the api after the changes
                            all the schedules are mapped to objects that are compatible with the api
                            incase of an error aborts the submission and denotes which schedule has an error*/
                            after.submit = function() {
                                /*tries to map the scheds to data shape that is api compatible 
                                and aborts on error*/
                                try {
                                    // trying to map the schedules as payload
                                    sched_payload = after.scheds.map(function(s) {
                                        try {
                                            return s.submit_data();
                                        } catch (sched) {
                                            // submit_data can throw error that is schedule which has it
                                            throw sched
                                        }
                                    })
                                    var payload = {
                                        serial: after.serial,
                                        scheds: sched_payload
                                    }
                                    console.log("Now ready to submit device details..");
                                    console.log(payload);
                                } catch (e) {
                                    // any one schedule that has error will be caught here
                                    console.error("Error in schedule : " + e.title);
                                    console.error(e.err.txt);
                                }

                            }
                        }
                    })
                    // getting the device details from the device serial
                srvRefactor($scope).get_object_from_api(function() {
                    return srvApi.get_device_schedules($routeParams.serial)
                }, function() {
                    console.error("Failed to get device schedules");
                }, "dd")
            }

        }
    }).directive("listOfSchedules", function() {
        return {
            restrict: "E",
            scope: {
                list: "=",
                rmaps: "<"
            },
            templateUrl: "/templates/list-schedules.html",
            controller: function($scope, ) {
                $scope.selcSchedule = null;
                var strTm_to_tmpick = function(strTm) {
                    // When on the GUI the string time received from the api does not work unless its disintegrated 
                    // api time format : 03:00 AM
                    // for the user to select and change this has to be disintegrated to hr, mn, mr format
                    // mr : meridiem
                    const pattern = /^([\d]{2}):([\d]{2}) (AM|PM)$/;
                    if (!pattern.test(strTm)) {
                        console.error("time strin in invalid format: " + strTm);
                        return null;
                    }
                    var result = strTm.match(pattern);
                    result.shift();
                    return {
                        hr: result[0],
                        mn: result[1],
                        mr: result[2],
                        format_time: function() {
                            return this.hr + ":" + this.mn + " " + this.mr;
                        }
                    }
                }
                var schedule = function(sch, index, isNewSched) {
                    /*constructor function for a new view model schedule
                    sch         : schedule as received from the api 
                    index       : index of the schedule in the list
                    isNewSched  : flag to denote if the schedule being inserted is new*/
                    sch.title = (sch.primary == true ? "Primary: " : "Exception: ") + index.toString();
                    sch.onTm = strTm_to_tmpick(sch.on);
                    sch.offTm = strTm_to_tmpick(sch.off);
                    /*Schedule ids:[] is just ids of the relays that the schedule applies to 
                    here we include the definition of each of the ids from rmaps
                    sel     : flag to denote if the relay id is selected, for new schedule this is selected*/
                    sch.ids = $scope.rmaps.map(function(m) {
                        var fltIds = sch.ids.filter(x => x == m.rid);
                        return {
                            id: m.rid,
                            text: m.defn,
                            sel: isNewSched == true ? false : fltIds.length > 0,
                            toggle_sel: function() {
                                if (sch.primary == false) {
                                    // incase of primary schedules ids cannot be deselected
                                    this.sel = !this.sel;
                                };
                            }
                        }
                    })
                    sch.submit_data = function(sch) {
                        /*Closure that emits a function that can prepare for submit. but incase the schedule has error will throw an error*/
                        return function() {
                            /*This is the reverse of extended model, when submitting the schedule it compacts the data shape to its bare necessities*/
                            if (sch.has_error() == false) {
                                return {
                                    on: this.onTm.format_time(),
                                    off: this.offTm.format_time(),
                                    ids: this.ids.filter(x => x.sel == true).map(x => x.id)
                                }
                            } else {
                                let idxInvalSched = $scope.list.findIndex(x => x.title == sch.title);
                                $scope.selcSchedule = $scope.list[idxInvalSched]
                                throw sch; //we throw back the schedule itself that has the error on it 
                            }
                        }

                    }(sch);
                    sch.err = null;
                    sch.has_error = function(sch) {
                        /*closure used to emit a function that can test if schedule has error before submit
                        atleast one relay node selected for non primary schedules and all relay nodes to be selected for primary schedule*/
                        return function() {
                            /*Incase not primary schedule atleast one relay node is to be selected while when primary schedule we need all the nodes to be selected */
                            var result = false;
                            result = sch.primary != true ? !(sch.ids.filter(x => x.sel == true).length > 0) : !(sch.ids.filter(x => x.sel == true).length == sch.ids.length);
                            /*Incase the schedule has error: it would be stamped on the schedule*/
                            sch.err = result == true ? { txt: "Schedules need atleast one relay node selected" } : null
                            return result;
                        }
                    }(sch);
                    return sch
                }
                $scope.del_schedule = function(schedTitle) {
                    /*using the title of the schedule this can remove a specific schedule from the list*/
                    var indexToDel = -1;
                    $scope.list.forEach(function(el, index) {
                        if (el.title == schedTitle) {
                            indexToDel = index;
                        }
                    });
                    if (indexToDel >= 0) {
                        $scope.list.splice(indexToDel, 1);
                        $scope.selcSchedule = indexToDel < ($scope.list.length - 1) ? $scope.list[indexToDel] : $scope.list[$scope.list.length - 1];
                        return
                    } else {
                        console.warn("Index to delete is invalid, not deleting any schedule");
                    }
                }
                $scope.add_schedule = function() {
                    /*adds a new schedule to the list of schedules 
                    newly added schedule will be a non primary schedule, default time, and no relay nodes selected*/
                    $scope.list.push(schedule({
                        on: "01:00 PM",
                        off: "01:00 AM",
                        primary: false,
                        ids: [],
                    }, $scope.list.length, true));
                    $scope.selcSchedule = $scope.list[$scope.list.length - 1];
                }
                $scope.$watch("list", function(after, before) {
                    // Once after the list is populated each of the schedule object gets a facelift 
                    // from the view model perspective schedules are modeled different from how they are received from the api
                    if (after !== null && after !== undefined) {
                        after.forEach(function(sch, index) {
                            schedule(sch, index, false);
                        });
                        if (after.length > 0) {
                            // default selecting the first schedule 
                            $scope.selcSchedule = $scope.list[0];
                        }
                        $scope.jsonData = JSON.stringify(after, null, 2)
                    }
                });


            }

        }
    })
})()