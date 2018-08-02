package controllers

import (
	"net/http"
	"github.com/schoolwheels/safestopclient/models"
	"github.com/gorilla/mux"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
)

type AppController struct {
	*ControllerBase
}

func (c *AppController) Register() {

	//templates
	c.addTemplate("index", "index.html", "default.html")
	c.addTemplate("check_availability", "check_availability.html", "default.html")


	//actions
	c.addRouteWithPrefix("/", c.indexAction)
	c.addRouteWithPrefix("/check_availability", c.checkAvailabilityAction)
	c.addRouteWithPrefix("/change_locale/{locale}", c.changeLocaleAction)


}

type dashData struct {
	CurrentUserEmail string
}


func (c *AppController) indexAction(w http.ResponseWriter, r *http.Request) {
	session, _ :=  c.SessionStore.Get(r, "auth")
	email := session.Values["current_user_email"]
	if email != nil {
		http.Redirect(w, r, r.URL.Host+"/dashboard", http.StatusFound)
	}
	c.render(w, r, "index", nil)
}

func (c *AppController) changeLocaleAction(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	session, err:= c.SessionStore.Get(r, "auth")
	session.Values["locale"] = vars["locale"]
	err = session.Save(r, w)


	if err != nil {
		//http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		//return
	}


	http.Redirect(w, r,"/login", http.StatusFound)
}




//CHECK AVAILABILITY
type checkAvailabilityData struct {
	PostalCode string
	Country string
}

func (c *AppController) checkAvailabilityAction(w http.ResponseWriter, r *http.Request) {

	if(r.FormValue("format") == "json"){
		available_jurisdictions := models.JurisdictionOptions{}
		available_jurisdictions.AuthInfo = validateToken(r.FormValue("token"))
		available_jurisdictions.AuthInfo.RedirectToLogin = false

		postal_code := r.FormValue("postal_code")
		pcr := models.PostalCodeReferenceForPostalCode(postal_code)
		if(pcr != nil){
			s := models.StateForAbbreviation(pcr.StateCode)
			if(s != nil){
				models.AvailableJurisdictionsForState(&available_jurisdictions, s.Id)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(structToJson(available_jurisdictions))

	} else{

		data := checkAvailabilityData{PostalCode: r.FormValue("postal_code"), Country: "US"}
		if(viper.GetString("domain") == "safestopapp.ca"){
			data.Country = "CA"
		}

		c.render(w, r, "check_availability", data)
	}
}



























// Redirects

func (c *AppController) redirectToJoinIfNotALoggedIn(w http.ResponseWriter, r *http.Request) {
	session, _ :=  c.SessionStore.Get(r, "auth")
	email := session.Values["current_user_email"]
	if email == nil {
		http.Redirect(w, r, r.URL.Host+"/join", http.StatusFound)
	}
}

func (c *AppController) redirectToLoginIfNotLoggedIn(w http.ResponseWriter, r *http.Request) {
	session, _ :=  c.SessionStore.Get(r, "auth")
	email := session.Values["current_user_email"]
	if email == nil {
		http.Redirect(w, r, r.URL.Host+"/join", http.StatusFound)
	}
}

// Helpers

func (c *AppController) getCurrentUserEmail(r *http.Request) string {
	session, _ :=  c.SessionStore.Get(r, "auth")
	email := session.Values["current_user_email"]
	return email.(string)
}

func (c *AppController) getCurrentUser(r *http.Request) *models.User {
	email := c.getCurrentUserEmail(r)
	user := models.FindUserByEmail(email)
	return user
}



func validateToken(token string) models.AuthInfo {
	a := models.AuthInfo{}
	u := models.FindUserByToken(token)
	if(u != nil){
		a.User = u
		a.TokenValid = true
	}
	return a
}

func structToJson(data interface{}) []byte{
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	} else{
		return b
	}
}