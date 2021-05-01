(function(){
    angular.module("luminapp").controller("schedCtrl", function($scope, $routeParams,srvApi, $route){
        $scope.wait = false;
        var origSchedules = JSON.stringify({}) //to start with its the imperession of an empty object
        $scope.$watch("schedules", function(after, before){
            if (after!==null && after!==undefined){
                // also make a json string for later comparison
                // after.forEach(function(el,index){
                //     el.editing = false;
                // })
                origSchedules = JSON.stringify(after)
            }
        })
        srvApi.get_device_schedules($routeParams.serial).then(function(data){
            console.log("We have received the schedules for the device: " +$routeParams.serial)
            console.table(data)
            $scope.schedules = data;
        }) //TODO: implement the error function too .. 
        $scope.submit = function(){
            console.log(JSON.stringify($scope.schedules))
            console.log(origSchedules)
            if (JSON.stringify($scope.schedules) !== origSchedules) {
                // Here we need to check to see if the original list is necessary before we push the changes.
                // remember that this patch operation would also involve the data broker sending a push notification to the device on the ground
                console.log("we are about to submit changes to the schedules")
                srvApi.patch_device_schedules($routeParams.serial,$scope.schedules).then(function(){
                    console.log("Success.. we have patched the schedules on the cloud");
                    $route.reload();
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
            $scope.schedules =$scope.schedules.filter(function(item,idx){
                return idx!=index;
            })
        }
        // Now from dummy schedules we switch to real schedules
        // $scope.schedules =[
        //     {"on":"06:30 PM", "off":"06:00 AM", "ids":["IN1","IN2","IN3","IN4"], "primary":true},
        //     {"on":"04:30 PM", "off":"06:29 AM", "ids":["IN1","IN4"], "primary":false},
        //     {"on":"02:30 AM", "off":"02:40 AM", "ids":["IN1","IN2"], "primary":false}
        // ]
    })
})()