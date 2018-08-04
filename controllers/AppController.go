package controllers

import (
	"net/http"
	"github.com/schoolwheels/safestopclient/models"
	"strings"
	"reflect"
	"github.com/gorilla/mux"
	"strconv"
)

type AppController struct {
	*ControllerBase
}

func (c *AppController) Register() {

	//templates
	c.addTemplate("index", "index.html", "app.html")
	c.addTemplate("check_availability", "check_availability.html", "default.html")
	c.addTemplate("account", "account.html", "app.html")
	c.addTemplate( "language", "language.html", "app.html")
	c.addTemplate( "activate", "activate.html", "default.html")

	//actions
	c.addRouteWithPrefix("/", c.IndexAction)
	c.addRouteWithPrefix("/check_availability", c.CheckAvailabilityAction)
	c.addRouteWithPrefix("/change_locale/{locale}", c.ChangeLocaleAction)
	c.addRouteWithPrefix("/account", c.AccountAction)
	c.addRouteWithPrefix( "/language", c.LanguageAction)
	c.addRouteWithPrefix("/activate/{jurisdiction_id}", c.ActivateAction)
}

type dashData struct {
	CurrentUserEmail string
}

func (c *AppController) IndexAction(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Token string
		Email string
	} {
		r.FormValue("token"),
		r.FormValue("email"),
	}


	session, _ :=  c.SessionStore.Get(r, "auth")
	email := session.Values["current_user_email"]
	if email != nil {
		http.Redirect(w, r, r.URL.Host+"/dashboard", http.StatusFound)
	}
	c.render(w, r, "index", data)
}

func (c *AppController) AccountAction(w http.ResponseWriter, r *http.Request) {

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
	}

	u := models.FindUser(uid)

	data := struct {
		JurisdictionCount int
	}{
		models.JurisdictionCountForUser(u, c.PermissionGroups),
	}



	c.render(w, r, "account", data)
}



func (c *AppController) CheckAvailabilityAction(w http.ResponseWriter, r *http.Request) {


	if r.Method == "GET" {

		data := struct {
			PostalCode string
			Jurisdictions *models.JurisdictionOptions
			JurisdictionCount int
		} {
			r.FormValue("postal_code"),
			&models.JurisdictionOptions{},
			-1,
		}

		c.render(w, r, "check_availability", data)

	} else {

		postal_code := r.FormValue("postal_code")
		jurisdictions := &models.JurisdictionOptions{}

		pcr := models.PostalCodeReferenceForPostalCode(postal_code)
		if(pcr != nil){
			s := models.StateForAbbreviation(pcr.StateCode)
			if(s != nil){
				jurisdictions = models.AvailableJurisdictionsForState(s.Id)
			}
		}

		data := struct {
			PostalCode string
			Jurisdictions *models.JurisdictionOptions
			JurisdictionCount int
		} {
			r.FormValue("postal_code"),
			jurisdictions,
			len(jurisdictions.Jurisdictions),
		}

		c.render(w, r, "check_availability", data)

	}

}









func (c *AppController) ActivateAction(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)


	if vars["jurisdiction_id"] != ""{

		id, err := strconv.Atoi(vars["jurisdiction_id"])
		if err != nil {
			//REDIRECT
		}

		data := struct {
			Jurisdiction interface{}
			PostalCode string
		} {
			models.ActivateJurisdiction(id),
			"29483",
		}
		c.render(w, r, "activate", data)

	} else {
		//REDIRECT
	}

}

func (c *AppController) ChangeLocaleAction(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	session, err:= c.SessionStore.Get(r, "auth")
	session.Values["locale"] = vars["locale"]
	err = session.Save(r, w)
	if err != nil {
		//http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		//return
	}
	http.Redirect(w, r,"/language", http.StatusFound)
}


func (c *AppController) LanguageAction(w http.ResponseWriter, r *http.Request) {

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
	}


	c.render(w, r, "language", nil)
}


func (c *AppController) MapAction(w http.ResponseWriter, r *http.Request) {
	session, _ :=  c.SessionStore.Get(r, "auth")
	email := session.Values["current_user_email"]
	if email != nil {
		http.Redirect(w, r, r.URL.Host+"/dashboard", http.StatusFound)
	}

	c.render(w, r, "map", nil)
}

func (c *AppController) FaqAction(w http.ResponseWriter, r *http.Request) {
	session, _ :=  c.SessionStore.Get(r, "auth")
	email := session.Values["current_user_email"]
	if email != nil {
		http.Redirect(w, r, r.URL.Host+"/dashboard", http.StatusFound)
	}

	c.render(w, r, "faq", nil)
}


// Redirects
//----------------------------------------------------------------------------------------------------------------------

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


// addAction requires you to have a view named <action>.html and a method func (c *AppController) <Action>Action(http.ResponseWriter, *http.Request)
func (c *AppController) addAction(action string){
	//TODO: determine if this can be moved to ControllerBase and if c can just be cast to the correct type.
	//fmt.Println(strings.Title(action)+"Action")
	c.addTemplateApp(action)
	c.Router.HandleFunc(c.RoutePrefix+"/"+action, reflect.ValueOf(c).MethodByName(strings.Title(action)+"Action").Interface().(func(http.ResponseWriter, *http.Request)))
}