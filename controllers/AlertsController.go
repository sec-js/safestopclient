package controllers

import (
	"github.com/schoolwheels/safestopclient/models"
	"net/http"
	"strings"
)

type AlertsController struct {
	*ControllerBase
}

func (c *AlertsController) Register() {

	//templates
	c.addTemplate("alerts_dashboard", "alerts_dashboard.html", "alerts_dashboard.html")
	c.addTemplate("alert_jurisdictions", "alert_jurisdictions.html", "alert_jurisdictions.html")
	c.addTemplate("alert_buses", "alert_buses.html", "alert_buses.html")
	c.addTemplate("alert_routes", "alert_routes.html", "alert_routes.html")
	c.addTemplate("create_alert", "create_alert.html", "create_alert.html")


	//actions
	c.addRouteWithPrefix("/alerts_dashboard", c.AlertsDashboardAction)
	c.addRouteWithPrefix("/alert_jurisdictions", c.AlertJurisdictionsAction)
	c.addRouteWithPrefix("/alert_buses", c.AlertBusesAction)
	c.addRouteWithPrefix("/alert_routes", c.AlertRoutesAction)
	c.addRouteWithPrefix("/create_alert", c.CreateAlertAction)
	c.addRouteWithPrefix("/process_create_alert", c.ProcessCreateAlertAction)

}


func (c *AlertsController) AlertsDashboardAction(w http.ResponseWriter, r *http.Request) {

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	u := models.FindUser(uid)

	if(models.UserCanSendAlerts(u) == false){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	data := struct {
		Email string
	} {
		u.Email,
	}

	c.render(w, r, "alerts_dashboard", data)
}


func (c *AlertsController) AlertJurisdictionsAction(w http.ResponseWriter, r *http.Request) {

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	u := models.FindUser(uid)
	if(models.UserCanSendAlerts(u) == false){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	data := struct {
		Email string
		Jurisdictions []models.AlertsJurisdiction
		Search string
	} {
		u.Email,
		*models.AlertsJurisdictionsForUser(u.Id, r.FormValue("search")),
		r.FormValue("search"),
	}


	c.render(w, r, "alert_jurisdictions", data)

}

func (c *AlertsController) AlertBusesAction(w http.ResponseWriter, r *http.Request) {

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	u := models.FindUser(uid)
	if(models.UserCanSendAlerts(u) == false){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	data := struct {
		Email string
		Buses []models.AlertsBus
		Search string
	} {
		u.Email,
		*models.AlertsBusesForUser(u.Id, r.FormValue("search")),
		r.FormValue("search"),
	}

	c.render(w, r, "alert_buses", data)
}




func (c *AlertsController) AlertRoutesAction(w http.ResponseWriter, r *http.Request) {

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	u := models.FindUser(uid)
	if(models.UserCanSendAlerts(u) == false){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	data := struct {
		Email string
		Routes []models.AlertsRoute
		Search string
	} {
		u.Email,
		*models.AlertsRoutesForUser(u.Id, r.FormValue("search")),
		r.FormValue("search"),
	}

	c.render(w, r, "alert_routes", data)
}





func (c *AlertsController) CreateAlertAction(w http.ResponseWriter, r *http.Request) {

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	u := models.FindUser(uid)
	if(models.UserCanSendAlerts(u) == false){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	jurisdiction_names := strings.Split(r.FormValue("jurisdiction_names"), ",")
	bus_names := strings.Split(r.FormValue("bus_names"), ",")
	route_names := strings.Split(r.FormValue("route_names"), ",")
	stop_names := strings.Split(r.FormValue("stop_names"), ",")


	data := struct {
		Email string
		BusIds string
		BusNames string
		BusNamesArray []string
		RouteIds string
		RouteNames string
		RouteNamesArray []string
		StopIds string
		StopNames string
		StopNamesArray []string
		JurisdictionIds string
		JurisdictionNames string
		JurisdictionNamesArray []string
	} {
		u.Email,
		r.FormValue("bus_ids"),
		r.FormValue("bus_names"),
		bus_names,
		r.FormValue("route_ids"),
		r.FormValue("route_names"),
		route_names,
		r.FormValue("stop_ids"),
		r.FormValue("stop_names"),
		stop_names,
		r.FormValue("jurisdiction_ids"),
		r.FormValue("jurisdiction_names"),
		jurisdiction_names,
	}

	c.render(w, r, "create_alert", data)

}




func (c *AlertsController) ProcessCreateAlertAction(w http.ResponseWriter, r *http.Request) {

	uid := currentUserId(c.ControllerBase, r)
	if (uid == 0) {
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	u := models.FindUser(uid)
	if(models.UserCanSendAlerts(u) == false){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	r.ParseForm();
	if len(r.FormValue("jurisdiction_ids")) > 0 {
		jurisdiction_alert_inserted := models.InsertAlerts(
			u,
			r.FormValue("jurisdiction_ids"),
			r.FormValue("priority"),
			r.FormValue("start_date"),
			r.FormValue("end_date"),
			r.FormValue("text"),
			"jurisdiction",
		)

		if jurisdiction_alert_inserted == false {
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
		}

		if jurisdiction_alert_inserted == true && r.FormValue("push_notification") != "" {
			devices := models.DevicesForJurisdictions(r.FormValue("jurisdiction_ids"))
			models.SendPushNotification(*devices, r.FormValue("text"))
		}



	}

	if len(r.FormValue("bus_ids")) > 0 {
		bus_alert_inserted := models.InsertAlerts(
			u,
			r.FormValue("bus_ids"),
			r.FormValue("priority"),
			r.FormValue("start_date"),
			r.FormValue("end_date"),
			r.FormValue("text"),
			"bus",
		)

		if bus_alert_inserted == false {
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
		}

		if bus_alert_inserted == true && r.FormValue("push_notification") != "" {
			devices := models.DevicesForBusIds(r.FormValue("bus_ids"))
			models.SendPushNotification(*devices, r.FormValue("text"))
		}
	}

	if len(r.FormValue("route_ids")) > 0 || len(r.FormValue("stop_ids")) > 0 {

		route_alert_inserted := false
		if len(r.FormValue("route_ids")) > 0 {
			route_alert_inserted = models.InsertAlerts(
				u,
				r.FormValue("route_ids"),
				r.FormValue("priority"),
				r.FormValue("start_date"),
				r.FormValue("end_date"),
				r.FormValue("text"),
				"bus_route",
			)
		}

		stop_alert_inserted := false
		if len(r.FormValue("stop_ids")) > 0 {
			route_alert_inserted = models.InsertAlerts(
				u,
				r.FormValue("stop_ids"),
				r.FormValue("priority"),
				r.FormValue("start_date"),
				r.FormValue("end_date"),
				r.FormValue("text"),
				"bus_route_stop",
			)
		}

		if route_alert_inserted == false && stop_alert_inserted == false {
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
		}

		if (route_alert_inserted == true || stop_alert_inserted == true) && r.FormValue("push_notification") != "" {
			devices := models.DevicesForRouteAndStopIds(r.FormValue("route_ids"), r.FormValue("stop_ids"))
			models.SendPushNotification(*devices, r.FormValue("text"))
		}

	}

	setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "your_alert_has_been_submitted", "")), c.BootstrapAlertClass.Info)
	http.Redirect(w, r, r.URL.Host+"/alerts_dashboard", http.StatusFound)
}
