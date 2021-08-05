(function(){
    angular.module("luminapp").directive("accTable", function(srvApi,srvRefactor,$route){
        return {
            restrict:"EA",
            replace:false,
            scope:false,
            transclude:true,
            templateUrl:"/templates/acc-table.html",
            compile :function(elem,attrs,transcludeFn){
                return {
                    post: function(scope, elem, attrs){
                        scope.locHide = attrs.locHide!=undefined ? true:false
                        scope.nameHide= attrs.nameHide !=undefined? true:false
                    }
                }
            },
            controller :function($scope){
                $scope.listOfAccs = [];
                srvRefactor($scope).get_list_from_api(function(){
                    // return srvApi.mock_api_fail(403)
                    return srvApi.list_accs()
                },function(el,index){
                    el.del = false;
                    el.mark_for_delete = function(){
                        el.del = !el.del;
                    }
                },function(){
                    console.error("Failed to download the list of accounts")
                },"listOfAccs")
                // this submit function runs when changes need to be submitted 
                $scope.submit =srvRefactor($scope).submit_list_changes("listOfAccs", function(el){
                    return el.del == true
                }, function(el){
                    return srvApi.remove_acc(el.email)
                }, function(){
                    console.error("Failed to remove user accounts");
                    $route.reload();
                })
            }
        }
    }).directive("deviceTable", function(srvRefactor,srvApi, $route, $routeParams){
        /*Used to enlist the devices of a user
        But when under admin login devices can be edited while when under user login then can be only viewed*/ 
        return {
            restrict:"EA",
            replace:false,
            scope:false,
            transclude:true,
            templateUrl:"/templates/device-table.html",
            controller : function($scope, $attrs){
                // Device lists under administration are editable 
                // while under user login will be viewable only
                $scope.editable = $attrs.edit ? ($attrs.edit =='true') : false
                console.log("Device table is editable: "+ $scope.editable);
                $scope.devices = [];
                if ($scope.authInfo) {
                    srvRefactor($scope).get_list_from_api(function(){
                        return srvApi.get_user_devices($routeParams.email);
                    },function(el,index){
                        // for each function needs no change
                        if ($scope.editable==true){
                            // only if the table is editable 
                            // else this function would be blank
                            el.black = false; // since all the devices will be from the registry list 
                            el.modified = false;
                            el.toggle_lock= function(){
                                if (el.black==false){
                                    // if the device is already blacked there is no point in locking /unlocking
                                    console.log("now toggling the lock status for the device")
                                    el.lock = !el.lock;
                                    el.modified = !el.modified;
                                }
                            }
                            // device black list is separate
                            el.toggle_black = function(){
                                // toggles the black status
                                el.black =!el.black;
                                // if its black listed only then its marked as modified
                                el.modified = el.black;
                            }
                        }
                    },function(){
                        console.error("Failed to download the list of accounts")
                    },"devices")
                    // below is to configure list submit function 
                    $scope.submit = srvRefactor($scope).submit_list_changes("devices", function(el){
                            return el.modified ==true;
                        }, function(el){
                            if (el.modified==true){
                                if (el.black ==true) {
                                    // if its blacklisted then its st fwd
                                    srvApi.blacklist_device(el.serial,true).then(function(){
                                        return srvApi.remove_luminreg(el.serial)
                                    }, function(err){
                                        return srvApi.fail_request(err)
                                    })
                                } else {
                                    // If the device is modified and not blacklisted then the lock status has been changed
                                    return srvApi.lock_device(el.serial,el.lock)
                                }
                            }
                        }, function(){
                            $route.reload();
                        })
                }else{
                    console.error("Failed to get authentication info");
                }
            }
        }
    }).directive("blackList", function(srvRefactor,srvApi){
        // A directive that lets you edit the blacklist
        // used by admins to white list devices from the ban
        return {
            restrict:"EA",
            replace:false,
            scope:false,
            transclude:true,
            templateUrl:"/templates/black-list.html",
            controller : function($scope){
                $scope.devices = [];
                srvRefactor($scope).get_list_from_api(function(){
                    return srvApi.get_device_blacklist();
                },function(el,index){
                    el.blacked = true; //since its the black list we are querying 
                    el.toggle = function(){
                        el.blacked = !el.blacked;
                    }
                },function(){
                    console.error("Failed to download the list of accounts")
                },"devices")
                // Now here we customize the submit function
                $scope.submit = srvRefactor($scope).submit_list_changes("devices", function(el){
                    return el.blacked ==false; //those that are removed from the blacklisting
                }, function(el){
                    // since this is used to only white list the devices 
                    return srvApi.blacklist_device(el.serial, false);
                }, function(){
                    $route.reload();
                })
            }
        }
    })
})()   