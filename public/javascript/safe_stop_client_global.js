/**
 * Created by adamcook on 4/27/17.
 */


function isNativeiOSApp() {
    return /SafeStop-iOS\/[0-9\.]+$/.test(navigator.userAgent);
}

function isNativeAndroidApp() {
    return /SafeStop-Android\/[0-9\.]+$/.test(navigator.userAgent);
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
    // alert("did_register_for_notifications_callback("+data+")");
    d = data.split("::");
    notification_token = d[0];
    device_platform = d[1];
    api_url = '/apns';

    if(notification_token.length > 0)
    {
        // alert("Registering "+device_platform+" device for APNS registrations with device_token: " + notification_token+" " +api_url);

        $.post(api_url, {
            device_token: notification_token,
            device_platform: device_platform
        }, function(response){
            //alert("Registered");
        }, "json");

    }
}

function did_register_for_notifications_with_user_callback(data)
{
    // alert("did_register_for_notifications_with_user_callback");
    d = data.split("::");
    notification_token = d[0];
    device_platform = d[1];
    token = window.localStorage.getItem("token");
    api_url = '/apns_user';

    if(notification_token.length > 0)
    {
        // alert("Registering "+device_platform+" device for APNS registrations with token: "+token+" and device_token: " + notification_token+" " +api_url);

        $.post(api_url, {
            device_token: notification_token,
            device_platform: device_platform,
            token: token
        }, function(response){
            //window.location = '//' + host_with_www + "/safe_stop_web_client/my_stops";
        }, "json");

    }
}

function gcm_callback(notification_token)
{
    device_platform = "Android";
    api_url = '/gcm';

    if(notification_token.length > 0)
    {
        //alert("Registering "+device_platform+" device for APNS registrations with token: "+token+" and device_token: " + notification_token+" " +api_url);

        $.post(api_url, {
            device_token: notification_token,
            device_platform: device_platform
        }, function(response){
            //alert("Registered");
        }, "json");

    }
}

function gcm_user_callback(notification_token)
{
    device_platform = "Android";
    token = window.localStorage.getItem("token");
    api_url = '/gcm_user';

    if(notification_token.length > 0)
    {
        // alert("Registering "+device_platform+" device for APNS registrations with token: "+token+" and device_token: " + notification_token+" " +api_url);

        $.post(api_url, {
            device_token: notification_token,
            device_platform: device_platform,
            token: token
        }, function(response){
            //window.location = '//' + host_with_www + "/safe_stop_web_client/my_stops";
        }, "json");

    }
}