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

// function isNativeiOSApp() {
//     return /(iPhone|iPod|iPad).*AppleWebKit(?!.*Safari)/i.test(navigator.userAgent);
// }

// function isNativeiOSApp() {
//     return /SafeStop-iOS\/[0-9\.]+$/.test(navigator.userAgent);
// }
//
// function isNativeAndroidApp() {
//     return /SafeStop-Android\/[0-9\.]+$/.test(navigator.userAgent);
// }
//
// function sendMessageToNative(json) {
//     try {
//         webkit.messageHandlers.callbackHandler.postMessage(json);
//     } catch (err) {
//         console.log('The native context does not exist yet');
//     }
// }

// function did_register_for_notifications_callback(data) {
//     d = data.split("::");
//     notification_token = d[0];
//     device_platform = d[1];
//     api_url = '//' + host_with_safestopapi + '/v1/apns';
//
//     if (notification_token.length > 0) {
//         $.post(api_url, {
//             device_token: notification_token,
//             device_platform: device_platform
//         }, function (response) {
//         }, "json");
//     }
// }
//
// function did_register_for_notifications_with_user_callback(data) {
//     d = data.split("::");
//     notification_token = d[0];
//     device_platform = d[1];
//     token = window.localStorage.getItem("token");
//     api_url = '//' + host_with_safestopapi + '/v1/apns_user';
//
//     if (notification_token.length > 0) {
//         $.post(api_url, {
//             device_token: notification_token,
//             device_platform: device_platform,
//             token: token
//         }, function (response) {
//             //window.location = '//' + host_with_www + "/safe_stop_web_client/my_stops";
//         }, "json");
//     }
// }
//
// function gcm_callback(notification_token) {
//     device_platform = "Android";
//     api_url = '//' + host_with_safestopapi + '/v1/gcm';
//
//     if (notification_token.length > 0) {
//         $.post(api_url, {
//             device_token: notification_token,
//             device_platform: device_platform
//         }, function (response) {
//         }, "json");
//     }
// }
//
// function gcm_user_callback(notification_token) {
//     device_platform = "Android";
//     token = window.localStorage.getItem("token");
//     api_url = '//' + host_with_safestopapi + '/v1/gcm_user';
//
//     if (notification_token.length > 0) {
//         $.post(api_url, {
//             device_token: notification_token,
//             device_platform: device_platform,
//             token: token
//         }, function (response) {
//             //window.location = '//' + host_with_www + "/safe_stop_web_client/my_stops";
//         }, "json");
//     }
// }


