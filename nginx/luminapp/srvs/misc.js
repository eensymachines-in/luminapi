(function(){
    angular.module("luminapp").service("srvRefactor", function($rootScope,$q, $route){
        /*this service aims to provide boilerplate functions or closures that can be called from controllers
        While controllers represent context, this here is the crux behind it
        All functions are here extension of scope of the then controller
        hence from the client code
        srvRefactor($scope).whatever_function() */ 
        return function(scope){
            // this makes the submit function for the buttons to call on
            // this assumes you have a list on top and items marked for change can be sent to srvApi on call
            this.submit_list_changes = function(listVarName,filter,service_call,errModal_ackw){
                var promises = [];
                return function(){
                    scope.wait = true;
                    scope[listVarName].forEach(function(el){
                        if (filter(el) ==true){
                            promises.push(service_call(el))
                        }
                    });
                    $q.all(promises).then(function(data){
                        console.info("Success! - submitted list change data")
                        scope.wait = false;
                        $route.reload();
                    }, function(error){
                        scope.wait = false;
                        error.upon_exit  = function(){
                            scope.$apply(function(){
                               errModal_ackw()
                            })
                        }
                        $rootScope.err = error;
                    })
                }
            }
            this.get_list_from_api = function(service_call,forEachFn,errModal_ackw,listVarName){
                /*
                service_call : closure that returns the promise from srvApi call, please see this is not a function but a closure
                forEachFn: function(el, index){} that needs to be run after the list data is downloaded
                errModal_ackw: error downloading the data shows a modal, acknowledging the modal needs a function again
                listVarName: name of the variable that the data is assigned to
                */ 
                scope.wait =true;
                service_call().then(function(data){
                    // since this is a list we would want to run a function for each element when the data list is downloaded
                    // this function is to overriden from the calling directive / controller
                    scope.wait =false;
                    console.log("Downloaded list data")
                    console.table(data)
                    data.forEach(forEachFn);
                    scope[listVarName] = data;
                }, function(error){
                    scope.wait =false;
                    error.upon_exit  = function(){
                        scope.$apply(function(){
                           errModal_ackw()
                        })
                    }
                    $rootScope.err = error;
                })
            }
            return this;
        }
    })
})()