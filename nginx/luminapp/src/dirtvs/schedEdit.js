(function(){
    angular.module("luminapp").directive("schedEdit", function(){
        // denotes an editable schedule
        // aggregates 2 time-select and a couple of labels
        // this is suitable for any schedule whatsoevr
        return {
            restrict:"E",
            replace:true,
            scope:{
                sched:"=", // the schedule to which this directive binds to 
                // this schedule shape is the one from the api
            },
            templateUrl:"/templates/sched-edit.html",
            controller : function($scope){
                // Not much going in around here in this controller since its the template that is important here
                // this pust out a section of couple of time-selects and shuttles the data change back to the main controller
                console.log("Inside schedEdit controller");               
            }
        }
    })
})()