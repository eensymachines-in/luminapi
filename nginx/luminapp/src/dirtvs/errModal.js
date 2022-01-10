(function(){
    angular.module("luminapp").factory("anyModal",function($rootScope){
        /*Boilerplate code for all modals. Modals can use different templates, but this can run the common logic behind all the modals
        Shows the modal when the rootScope variable changes
        implements the show, hide and upon_hidden behaviour
        From the directive of the modal, make a new instance of this and watch from the controller while link in the element from link of the directive*/ 
        return function(scope, varName, modalId){
            this.link = function(elem,title_text,body_text) {
                // title_text : a function that will return what the title of the modal would be
                // body_text : a function that will return what the body of the modal would be
                scope.mod =  $(elem).children("#"+modalId);
                scope.mod.on('hidden.bs.modal', function(){
                    $rootScope[varName].upon_exit()
                    $rootScope[varName] = null;
                }) 
                scope.show = function(){
                    // $(mod).children('h5.modal-title').text(scope.err.status+" "+scope.err.statusText)
                    scope.mod.find('.modal-title-span').text(title_text())
                    scope.mod.find('.modal-body').text(body_text())
                    scope.mod.modal('show')
                }
                scope.hide = function(){
                    scope.mod.modal('hide')
                }
            }
            this.watch = function(){
                $rootScope.$watch(varName, function(after, before){
                    if (after) {
                        // scope.hide();
                        scope.show();
                    }
                })
            }
        }
    })
    .directive("errModal", function($rootScope,anyModal){
        return {
            restrict:"EA",
            replace:false,
            scope:{}, // now that the err is on the $rootScope the directive can use isolated scope
            templateUrl:"/templates/err-modal.html",
            link : function(scope, elem, attrs){
                // console.log("Now testing var name "+ scope.am.get_varName())
                scope.am.link(elem,function(){
                    return $rootScope.err.status?$rootScope.err.status+" "+$rootScope.err.statusText:$rootScope.err.statusText
                }, function(){
                    return $rootScope.err.message
                })
            },
            controller : function($scope){
                // err object resides on the rootScope for any of the directives to access it
                // $rootScope.err = null;
                $scope.am = new anyModal($scope, "err", "errModal")
                $scope.am.watch(); //this shall set up the watch in the context of the scope
            }

        }
    }).directive("successModal", function($rootScope,anyModal){
        return {
            restrict:"EA",
            replace:false,
            scope:{}, // now that the err is on the $rootScope the directive can use isolated scope
            templateUrl:"/templates/succ-modal.html",
            link:function(scope, elem, attrs){
                // console.log("Now testing var name "+ scope.am.get_varName())
                scope.am.link(elem,function(){
                    return $rootScope.success.title
                }, function(){
                    return $rootScope.success.message
                })
            },
            controller:function($scope){
                // err object resides on the rootScope for any of the directives to access it
                $scope.am =new anyModal($scope, "success", "succModal")
                $scope.am.watch(); //this shall set up the watch in the context of the scope
            }

        }
    })
})() 