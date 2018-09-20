//var safe_stop_red = #A12130;

Mustache.tags = ["{|", "|}"];

var is_uiwebview = /(iPhone|iPod|iPad).*AppleWebKit(?!.*Safari)/i.test(navigator.userAgent);

var red = '#e94d2f';
var green = '#06af4b';
var blue = '#008fa8';
var yellow = '#EEA925';

var progressValue = 0,
    inProgress,
    progressTimeout;


function progress() {
    progressValue -= 1;
    $(".progress-bar").css("width", progressValue + "%");
    progressTimeout = setTimeout(progress, 300)
}

function isBadIE() {
    if (navigator.appVersion.indexOf("MSIE 9") == -1 &&
        navigator.appVersion.indexOf("MSIE 8") == -1 &&
        navigator.appVersion.indexOf("MSIE 7") == -1 &&
        navigator.appVersion.indexOf("MSIE 6") == -1) {
        return false
    }
    return true;
}



var all_alerts = [];
var progress_value;
var in_progress = false;

function start_progress_bar(){
    //START PROGRESS BAR
    progress_value = 100;
    if (!in_progress) {
        refresh_messages();
        refresh_stops();
        refresh_scan_notifications();
        refresh_ad();
        progress();
        in_progress = !in_progress;
    }
}

function progress() {
    progress_value -= 1;
    $(".progress-bar").css("width", progress_value + "%");

    if(progress_value == 0){
        refresh_messages();
        refresh_stops();
        refresh_scan_notifications();
        refresh_ad();
        progress_value = 100;
    }
    progress_timeout = setTimeout(progress, 300)
}

function refresh_messages() {

    $.post('/api/alerts', {"gorilla.csrf.Token": $("input[name='gorilla.csrf.Token']").val() }, function (response) {

        $("#messages").empty();

        if (response.unread_messages != '0') {
            $("#message-filled").show();
            $("#message-close").hide();
            $("#message-empty").hide();
        }

        for (var x = 0; x < response.alerts.length; x++) {

            if (response.alerts[x].priority == "info") {
                $("#messages").append("<tr><td style='padding:4px 4px 4px 4px; font-size: 16pt; color:#5bc0de; vertical-align: top;'><i class='fa fa-exclamation-circle'></i></td><td style='padding:4px;'>" + response.alerts[x].text + "</td></tr>");
            }
            else if (response.alerts[x].priority == "alert") {
                $("#messages").append("<tr><td style='padding:4px 4px 4px 4px; font-size: 16pt; color:#d9534f; vertical-align: top;'><i class='fa fa-exclamation-triangle'></i></td><td style='padding:4px;'>" + response.alerts[x].text + "</td></tr>");
            }
            else if (response.alerts[x].priority == "audible") {
                $("#messages").append("<tr><td style='padding:4px 4px 4px 4px; font-size: 16pt; color:#FCD20A; vertical-align: top;'><i class='fa fa-clock-o'></i></td><td style='padding:4px;'>" + response.alerts[x].text + "</td></tr>");
            }
            else if (response.alerts[x].priority == "ad-offer") {
                $("#messages").append("<tr><td style='padding:4px 4px 4px 4px; font-size: 16pt; color:#4CB349; vertical-align: top;'><i class='fa fa-tag'></i></td><td style='padding:4px;'>" + response.alerts[x].text + "</td></tr>");
            }
            else if (response.alerts[x].priority == "ad-other") {
                $("#messages").append("<tr><td style='padding:4px 4px 4px 4px; font-size: 16pt; color:#FCD20A; vertical-align: top;'><i class='fa fa-star'></i></td><td style='padding:4px;'>" + response.alerts[x].text + "</td></tr>");
            }

            all_alerts.push(response.alerts[x].id);
        }
    });
    refreshMessagesTimeout = setTimeout(refresh_messages, 30000);
}

function set_viewed_alerts(){

    $.post('/api/set_viewed_alerts', {
        "gorilla.csrf.Token": $("input[name='gorilla.csrf.Token']").val(),
        alert_ids: all_alerts.join(',')
    }, function (response) {
        // console.log(response)
    });

    if($("#overlay").is(':visible'))
    {
        $('#overlay').hide();
    }
    else
    {
        $('#overlay').show();
    }

    if($("#message-empty").is(':visible') || $("#message-filled").is(':visible'))
    {
        $("#message-close").show();
        $("#message-empty").hide();
        $("#message-filled").hide();
    }
    else
    {
        $("#message-close").hide();
        $("#message-empty").show();
        $("#message-filled").hide();
    }
}

function refresh_ad() {
    $.post('/api/next_ad', {
        "gorilla.csrf.Token": $("input[name='gorilla.csrf.Token']").val(),
    }, function (response) {

        if (response.id > 0) {
            $("#ad").show();

            $("#ad-image").attr("src", response.url);
            $("#ad-link").attr("href", '/adclick/' + response.id);

            setTimeout(function(){

                if($("#my-stops")) {
                    $("#my-stops").css('height', $(window).height() - $(".nav-bar-bottom").height() - $(".nav-bar-top").height() + 'px');
                }

                if($("#map")){
                    $("#map").css('height', $(window).height() - $(".nav-bar-bottom").height() - $(".nav-bar-top").height() + 'px');
                }

            }, 1000);

        }
    });
}

function isNativeiOSApp() {
    return /SafeStop-iOS\/[0-9\.]+$/.test(navigator.userAgent);
}

function isNativeAndroidApp() {
    return /SafeStop-Android/.test(navigator.userAgent);
}

function sendMessageToNative(json) {
    try {
        webkit.messageHandlers.callbackHandler.postMessage(json);
    } catch(err) {
        console.log('The native context does not exist yet');
    }
}

function did_register_for_notifications_callback(data)
{
    d = data.split("::");
    notification_token = d[0];
    device_platform = d[1];
    if(notification_token.length > 0)
    {
        $.post('/api/register_for_push_notifications', {
            device_token: notification_token,
            device_platform: device_platform,
            "gorilla.csrf.Token": $("input[name='gorilla.csrf.Token']").val(),
        }, function(response){

        }, "json");
    }
}

function gcm_callback()
{
    var gcm_token = Android.getGCMToken();
    if (gcm_token && gcm_token.length > 0) {
        if (ntk && ntk.length > 0 && tk && tk.length > 0){
            $.post('/api/register_for_push_notifications', {
                device_token: gcm_token,
                device_platform: "Android",
                "gorilla.csrf.Token": $("input[name='gorilla.csrf.Token']").val(),
            }, function(response){
                window.localStorage.removeItem("gcm_token");
            }, "json");
        }
    }
}