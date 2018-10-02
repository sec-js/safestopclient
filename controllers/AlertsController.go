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



}


func (c *AlertsController) AlertsDashboardAction(w http.ResponseWriter, r *http.Request) {

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	u := models.FindUser(uid)


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


	data := struct {
		Email string
		Jurisdictions []models.AlertsJurisdiction
		Search string
	} {
		u.Email,
		*models.AlertsJurisdictionsForUser(u.Id, r.FormValue("search")),
		r.FormValue("search"),
	}


	if r.Method == "GET" {
		c.render(w, r, "alert_jurisdictions", data)
	} else {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)

	}

}

func (c *AlertsController) AlertBusesAction(w http.ResponseWriter, r *http.Request) {

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	u := models.FindUser(uid)

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


	data := struct {
		Email string
	} {
		u.Email,
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


	jurisdiction_names := strings.Split(r.FormValue("jurisdiction_names"), ",")
	bus_names := strings.Split(r.FormValue("bus_names"), ",")


	data := struct {
		Email string
		JurisdictionNames []string
		BusIds string
		BusNames []string
		RouteIds string
		StopIds string
		JurisdictionIds string
	} {
		u.Email,
		jurisdiction_names,
		r.FormValue("bus_ids"),
		bus_names,
		"",
		"",
		r.FormValue("jurisdiction_ids"),
	}

	c.render(w, r, "create_alert", data)

}