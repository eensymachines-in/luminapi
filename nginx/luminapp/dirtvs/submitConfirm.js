
(function(){
    angular.module("luminapp").directive("submitConfirm", function($sce){
        return {
            restrict:"EA",
            replace: true,
            templateUrl:"/templates/submit-confirm.html",
            scope:true,
            link : function(scope,elem,attrs){
                scope.btnLbl = attrs.btnLbl!= undefined ? $sce.trustAsHtml(attrs.btnLbl) : $sce.trustAsHtml("Submit")
            },
            controller: function($scope){
                $scope.upon_submit = function(){
                    console.log("About to submit the login credentials")
                    $scope.submit();
                }
            }

        }
    }).directive("popConfirm", function(){
        // diretive that when placed on a submit button can help you show a popup before submitting 
        return {
            restrict :"A",
            replace:false, 
            scope :false, 
            compile : function(tElem, tAttrs){
                return {
                    pre :function(scope, elem,attrs){
                        var btn = angular.element(elem).find('.btn');
                        btn.attr('data-trigger', 'click')
                        btn.attr('data-toggle', 'popover')
                        btn.attr('data-placement', 'top')
                        btn.attr('data-content', 'This will make <strong class="text-warning">permanent</strong> changes')
                        btn.attr('data-html', 'true')
                        btn.attr('title', '<span class="text-warning">Are you sure?</span>')
                    },
                    post :function(scope, elem ,attrs){
                        var popElem =angular.element(elem).find('.btn');
                        popElem.popover()
                        scope.upon_submit = function(){
                            // this will be emoty since all what we need is the click to trigger the popover
                            // please see we are overriding the action from submit-confirm 
                            // now the button no longer submits on click 
                        }
                        popElem.on('hidden.bs.popover', function(){
                            popElem.removeAttr('data-trigger')
                            scope.submit()
                        })
                    }
                }
            }
        }
        
    })
})()