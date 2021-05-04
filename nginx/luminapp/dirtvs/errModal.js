(function(){
    angular.module("luminapp").factory("srvModal", function($rootScope,$timeout,$sce){
        return function(scope){
            // this has all the generic stuff that modal require 
            this.link = function(elem, attrs,varName,title_text,body_text) {
                // title_text : a function that will return what the title of the modal would be
                // body_text : a function that will return what the body of the modal would be
                // call this from within the directive link function
                // this can link the modal with the template 
                mod =  $(elem).children("div.modal")
                mod.on('hidden.bs.modal', function(){
                    console.log(varName)
                    $rootScope[varName].upon_exit()
                    $rootScope[varName] = null; // the error is acknowledged and can be dismissed
                }) 
                scope.show = function(){
                    // $(mod).children('h5.modal-title').text(scope.err.status+" "+scope.err.statusText)
                    mod.find('.modal-title').text($sce.trustAsHtml(title_text()))
                    mod.find('.modal-body').text(body_text())
                    mod.modal('show')
                }
                scope.hide = function(){
                    mod.modal('hide')
                }
            }
            this.watch = function(varName){
                $rootScope.$watch(varName, function(after, before){
                    if (after!== null) {
                        if (before === null) {
                            scope.show();
                        }else if (after !== before){
                            scope.hide();
                            $timeout(function(){
                                scope.show();
                            },500)
                        }
                    }
                })
            }
            return this;
        }
        
    }).directive("errModal", function($rootScope,srvModal){
        return {
            restrict:"EA",
            replace:false,
            scope:{}, // now that the err is on the $rootScope the directive can use isolated scope
            templateUrl:"/templates/err-modal.html",
            link : function(scope, elem, attrs){
                srvModal(scope).link(elem, attrs, "err",function(){
                    return $rootScope.err.status?$rootScope.err.status+" "+$rootScope.err.statusText:$rootScope.err.statusText
                }, function(){
                    return $rootScope.err.message
                })
            },
            controller : function($scope){
                // err object resides on the rootScope for any of the directives to access it
                // $rootScope.err = null;
                srvModal($scope).watch("err"); //this shall set up the watch in the context of the scope
            }

        }
    }).directive("successModal", function($rootScope,srvModal){
        return {
            restrict:"EA",
            replace:false,
            scope:{}, // now that the err is on the $rootScope the directive can use isolated scope
            templateUrl:"/templates/err-modal.html",
            link:function(scope, elem, attrs){
                console.log("Now linking the success modal");
                srvModal(scope).link(elem, attrs, "err",function(){
                    return $rootScope.err.title
                }, function(){
                    return $rootScope.err.message
                })
            },
            controller:function($scope){
                // err object resides on the rootScope for any of the directives to access it
                $rootScope.err = null;
                srvModal($scope).watch("err"); //this shall set up the watch in the context of the scope
            }

        }
    })
})() 