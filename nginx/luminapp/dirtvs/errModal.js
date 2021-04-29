(function(){
    angular.module("luminapp").directive("errModal", function($rootScope, $timeout){
        return {
            restrict:"EA",
            replace:false,
            scope:{}, // now that the err is on the $rootScope the directive can use isolated scope
            templateUrl:"/templates/err-modal.html",
            link : function(scope, elem, attrs){
                mod =  $(elem).children("div.modal")
                mod.on('hidden.bs.modal', function(){
                    console.log(scope.err)
                    $rootScope.err.upon_exit()
                    $rootScope.err = null; // the error is acknowledged and can be dismissed
                }) 
                scope.show = function(){
                    // $(mod).children('h5.modal-title').text(scope.err.status+" "+scope.err.statusText)
                    mod.find('.modal-title').text($rootScope.err.status?$rootScope.err.status+" "+$rootScope.err.statusText:$rootScope.err.statusText)
                    mod.find('.modal-body').text($rootScope.err.message)
                    mod.modal('show')
                }
                scope.hide = function(){
                    mod.modal('hide')
                }
            },
            controller : function($scope){
                // err object resides on the rootScope for any of the directives to access it
                $rootScope.err = null;
                $rootScope.$watch("err", function(after, before){
                    if (after!== null) {
                        if (before === null) {
                            $scope.show();
                        }else if (after !== before){
                            $scope.hide();
                            $timeout(function(){
                                $scope.show();
                            },500)
                        }
                    }
                })
            }

        }
    })
})() 