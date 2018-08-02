//var safe_stop_red = #A12130;

Mustache.tags = ["{|", "|}"];



function safe_stop_get(path, data, successHandler, errorHandler){
    var token = localStorage.getItem("token");
    data["token"] = token;

    $.get(path, data, function(response){

         // alert(JSON.stringify(response));

        if(response.auth_info.token_valid == false && response.auth_info.redirect_to_login == true){
            window.location.href = '/login'
        }
        else if(response.error.msg != ""){
            if(errorHandler != undefined){
                errorHandler(response);
            }
        }
        else {
            successHandler(response);
        }
    })
}

function safe_stop_post(path, data, successHandler, errorHandler){
    var token = localStorage.getItem("token");
    data["token"] = token;

    $.post(path, data, function(response){

        // alert(JSON.stringify(response));

        if(response.auth_info.token_valid == false && response.auth_info.redirect_to_login == true){
            window.location.href = '/login'
        }
        else if(response.error.msg != ""){
            if(errorHandler != undefined){
                errorHandler(response);
            }
        }
        else {
            successHandler(response);
        }
    })
}