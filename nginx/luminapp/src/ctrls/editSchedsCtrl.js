(function() {
    angular.module("luminapp").directive("getDeviceScheds", function() {
        return {
            restrict: "A",
            scope: false,
            controller: function($scope, srvRefactor, $routeParams, srvApi) {
                srvRefactor($scope).get_object_from_api(function() {
                    return srvApi.get_device_schedules($routeParams.serial)
                }, function() {
                    console.error("Failed to get device schedules");
                }, "dd")
            }

        }
    }).directive("deviceDetails", function() {
        return {
            restrict: "A",
            scope: false,
            controller: function($scope, ) {
                $scope.$watch("dd", function(after, before) {
                    if (after !== null && after !== undefined) {
                        after.submit = function() {
                            try {
                                sched_payload = after.scheds.map(function(s) {
                                    try {
                                        return s.submit_data();
                                    } catch (e) {
                                        throw e
                                    }
                                })
                                var payload = {
                                    serial: after.serial,
                                    scheds: sched_payload
                                }
                                console.log("Now ready to submit device details..");
                                console.log(payload);
                            } catch (e) {
                                console.error("submit error: " + e)
                            }

                        }
                    }
                })
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
                var rmaps_to_ids = function(sch, boolNewSch) {
                    /*schedules have ids of the relays while the rmaps have the complete definition 
                    we need a object that is actually the view model by the cross multiplication of 2 */
                    return $scope.rmaps.map(function(m) {
                        var fltIds = sch.ids.filter(x => x == m.rid);
                        return {
                            id: m.rid,
                            text: m.defn,
                            sel: boolNewSch == true ? false : fltIds.length > 0,
                            toggle_sel: function() {
                                if (sch.primary == false) {
                                    // incase of primary schedules ids cannot be deselected
                                    this.sel = !this.sel;
                                };
                            }
                        }
                    })

                }
                var submit_sch_data = function(sch) {
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
                            throw "schedule: " + this.title + " has error. See if atleast one relay node is selected"
                        }
                    }

                }
                var schedule_has_err = function(sch) {
                    /*closure used to emit a function that can test if schedule has error before submit
                    atleast one relay node selected for non primary schedules and all relay nodes to be selected for primary schedule*/
                    return function() {
                        /*Incase not primary schedule atleast one relay node is to be selected while when primary schedule we need all the nodes to be selected */
                        if (sch.primary != true) {
                            return !(sch.ids.filter(x => x.sel == true).length > 0);
                        } else {
                            return !(sch.ids.filter(x => x.sel == true).length == sch.ids.length);
                        }
                    }
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
                    var sch = {
                        title: "Exception: " + ($scope.list.length).toString(),
                        onTm: strTm_to_tmpick("01:00 PM"),
                        offTm: strTm_to_tmpick("01:00 AM"),
                        primary: false,
                        ids: [],
                    }
                    sch.ids = rmaps_to_ids(sch, true);
                    sch.submit_data = submit_sch_data(sch);
                    sch.has_error = schedule_has_err(sch);
                    $scope.list.push(sch);
                    $scope.selcSchedule = $scope.list[$scope.list.length - 1];
                }
                $scope.$watch("list", function(after, before) {
                    // Once after the list is populated each of the schedule object gets a facelift 
                    // from the view model perspective schedules are modeled different from how they are received from the api
                    if (after !== null && after !== undefined) {
                        after.forEach(function(sch, index) {
                            sch.title = (sch.primary == true ? "Primary: " : "Exception: ") + index.toString();
                            sch.onTm = strTm_to_tmpick(sch.on);
                            sch.offTm = strTm_to_tmpick(sch.off);
                            sch.ids = rmaps_to_ids(sch, false);
                            sch.submit_data = submit_sch_data(sch);
                            sch.has_error = schedule_has_err(sch);
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