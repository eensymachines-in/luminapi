(function(){
    // Lets you select time from a group of 3 drop downs
    // whats customized is the format of the time that needs to be posted to the api
    // data shape of the time for the schedules is specific to the api
    // From the controllers above one can push in the time as string ex: '10:30 AM'
    // this here is split to individual components of hr:min AM/PM selectors
    // once the user has selected and made changes this will put this back to the required string format
    angular.module("luminapp").directive("timeSelect", function($timeout){
        return {
            restrict:"E",
            replace:false,
            scope:{
                tm :"=", // this is the time bound from top - string time
                title:"@",
                desc:"@"
            },
            templateUrl:"/templates/time-sel.html",
            controller : function($scope){
                // making objects that represent elements on the UI
                // this watch is needed only when the data flows from the top 
                // but when this directive re-assigns the the value to tm the watch need not run
                $scope.hrSelect  = {val: "", opts:[]};
                $scope.minSelect = {val: "", opts:[]};
                $scope.ampmSelect = {val:"", opts:["AM","PM"]}
                // ++++++++++ below 2 are convertor functions string <--> split format
                // this directive works on the split format while the time set from the controller is string "09:34 PM"
                internal = false;
                $scope.$watch("tm", function(after, before){
                    if (after){
                        if (!internal){
                            // when the call is internal we wouldnt to get into an update cycle
                            split1 = after.split(" ")
                            if (split1.length ==2) {
                                $scope.ampmSelect.val = split1[1];      //  getting the AM PM from the time
                                split2=split1[0].split(":")
                                if (split2.length ==2) {
                                    $scope.hrSelect.val = split2[0];    //  getting the hours from the time
                                    $scope.minSelect.val = split2[1]    //  getting the minutes from the time
                                }
                            }
                        }else{
                            console.log("time changed but internal, will not re-split")
                        }
                    } else  {
                        console.log("time changed, but will NOT split");
                        console.log("Debugging before/after values: "+ before +" " +after);
                    }
                })
                var fragment_changed  = function(after, before){
                    internal = true;
                    if (after && after!= before) {
                        $scope.tm =$scope.hrSelect.val+":"+$scope.minSelect.val+" "+$scope.ampmSelect.val;
                        console.log("time fragment changed..");
                    }
                    internal = false;
                }
                $scope.$watch("hrSelect.val", fragment_changed);
                $scope.$watch("minSelect.val", fragment_changed);
                $scope.$watch("ampmSelect.val", fragment_changed);

                // +++++++++++ filling out the options
                for (i=1;i<=12;i++){
                    // options for hour select are padded numbers from 1-12 
                    $scope.hrSelect.opts.push(String(i).padStart(2,'0'))
                }
                // this has to change once we have the $scope.tm changed 
                // ++++++++++++++++++++ Minute value calculations and object setup
                for (i=0;i<=59;i++){
                    // options for minute select are padded numbers from 0-59
                    $scope.minSelect.opts.push(String(i).padStart(2,'0'))
                }
                // ++++++++++++ setting preselects
                // If the value for the hr is not selected we default to the first value in theoptions
                if ($scope.hrSelect.val == "") {
                    $scope.hrSelect.val = $scope.hrSelect.opts[0];
                }
                if ($scope.minSelect.val ==""){
                    // Incase the minutes are not set we resort to default first value
                    $scope.minSelect.val = $scope.minSelect.opts[0];
                }
                if($scope.ampmSelect.val == "") {
                    // Incase the AM PM is not set we resort to opts[0]
                    $scope.ampmSelect.val = $scope.ampmSelect.opts[0]
                }
            }
        }
    })
})()