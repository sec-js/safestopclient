//var safe_stop_red = #A12130;

Mustache.tags = ["{|", "|}"];

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
