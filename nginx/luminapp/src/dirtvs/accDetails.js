(function(){
    angular.module("luminapp").directive("accDetails", function(emailPattern,passwdPattern){
        return {
            restrict:"EA",
            replace:false,
            scope:false,
            transclude:true,
            templateUrl:"/templates/acc-details.html",
            compile : function(tElem, tAttrs, transcludeFn){
                return {
                    pre: function(scope, elem, attrs){
                        
                    },
                    post:function(scope, elem, attrs){
                        scope.lblHide=attrs.lblHide!==undefined? true:false;
                        scope.roleHide=attrs.roleHide!==undefined?true:false;
                        scope.passwdHide = attrs.passwdHide!==undefined?true:false;
                        scope.emailHide = attrs.emailHide!==undefined?true:false;
                        scope.detailsHide = attrs.detailsHide!==undefined?true:false;
        
                        scope.roleEdit=attrs.roleEdit!==undefined?true:false;
                        scope.emailEdit=attrs.emailEdit!==undefined?true:false;
        
                        if (scope.roleHide == true)  {
                            delete scope.details.role;
                        }
                        if (scope.passwdHide == true)  {
                            delete scope.details.passwd;
                        }
                        if (scope.detailsHide == true)  {
                            delete scope.details.name;
                            delete scope.details.loc;
                            delete scope.details.phone;
                            delete scope.details.role;
                        }
                        if(scope.emailHide ==true){
                            delete scope.details.email;
                        }
                    }
                }
            },
            controller : function($scope){
                $scope.isEmailInvalid = false;
                $scope.isPassInvalid = false;
                $scope.$on("validate", function(evt,data){
                    if ($scope.emailEdit==true && $scope.emailHide ==false){
                        $scope.isEmailInvalid = !emailPattern.test($scope.details.email) || $scope.details.email =="";
                    }
                    if ($scope.passwdHide ==false){
                        $scope.isPassInvalid = !passwdPattern.test($scope.details.passwd) || $scope.details.passwd =="";
                    }
                })
                $scope.details = {
                    email:"",
                    passwd:"",
                    name :"",
                    loc:"",
                    phone:"",
                    role: 0
                }
            }

        }
    })
})() 