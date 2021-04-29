(function(){
    angular.module("luminapp").directive("socialFooter",function(){
        return {
            restrict :"E",
            replace:true,
            templateUrl:"/templates/social-footer.html"
        }
    })
})()