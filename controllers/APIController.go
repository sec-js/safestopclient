package controllers

import (
	"net/http"
	"github.com/schoolwheels/safestopclient/models"
	"strconv"
)

type APIController struct {
	*ControllerBase
}

func (c *APIController) Register() {

	c.RoutePrefix = "/api"

	//actions
	c.addRouteWithPrefix("/version", c.versionAction)
	c.addRouteWithPrefix("/student_exists", c.StudentExistsAction)
	c.addRouteWithPrefix("/school_code_exists", c.SchoolCodeExistsAction)
	c.addRouteWithPrefix("/email_exists", c.EmailExistsAction)

	c.addRouteWithPrefix("/test", c.TestAction)


}

func (c *APIController) versionAction(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Version float64 `json:"version"`
	}{
		1.0,
	}

	c.renderJSON(data, w)
}















//http://ssc.local:8080/api/student_exists?sis_identifier=112408&jurisdiction_id%20=%2015
func (c *APIController) StudentExistsAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()

	sis_identifier := r.FormValue("sis_identifier")
	jurisdiction_id, err :=  strconv.Atoi(r.FormValue("jurisdiction_id"))
	if err != nil || sis_identifier == "" {
		v := struct {
			Valid bool `json:"valid"`
		} {
			false,
		}
		c.renderJSON(v, w)
	} else {
		v := struct {
			Valid bool `json:"valid"`
		} {
			models.StudentIdentifierExists(sis_identifier, jurisdiction_id),
		}

		c.renderJSON(v, w)
	}
}

//http://ssc.local:8080/api/school_code_exists?school_code=MACC48&jurisdiction_id=214
func (c *APIController) SchoolCodeExistsAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()

	school_code := r.FormValue("school_code")
	jurisdiction_id, err :=  strconv.Atoi(r.FormValue("jurisdiction_id"))
	if err != nil || school_code == "" {
		v := struct {
			Valid bool `json:"valid"`
		} {
			false,
		}
		c.renderJSON(v, w)
	} else {
		v := struct {
			Valid bool `json:"valid"`
		} {
			models.SchoolCodeExists(school_code, jurisdiction_id),
		}
		c.renderJSON(v, w)
	}
}

//http://ssc.local:8080/api/email_exists?user[email]=acook@ridesta.comfff
func (c *APIController) EmailExistsAction(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	v := struct {
		Valid bool `json:"valid"`
	} {
		!models.EmailExists(r.FormValue("user[email]")),
	}
	w.Header().Set("Content-Type", "application/json")
	c.renderJSON(v, w)
}



//http://ssc.local:8080/api/next_registration_ad?jurisdiction_id=172
func (c *APIController) TestAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c.renderJSON(models.ActivateJurisdiction(57), w)
}

