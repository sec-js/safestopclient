package controllers

import (
	"net/http"
	"fmt"
	"github.com/schoolwheels/safestopclient/models"
	"log"
	"os"
)

type AuthController struct {
	*ControllerBase
}

func init() {
	fmt.Print()
}

func (c *AuthController) Register() {

	//templates
	c.addTemplate("login", "login.html", "default.html")
	c.addTemplate("register", "register.html", "default.html")
	c.addTemplate("forgot_password", "forgot_password.html", "default.html")
	c.addTemplate("forgot_password_result", "forgot_password_result.html", "default.html")

	//actions
	c.addRouteWithPrefix("/login", c.loginAction)
	c.addRouteWithPrefix("/logout", c.logoutAction)
	c.addRouteWithPrefix("/register", c.registerAction)
	c.addRouteWithPrefix("/forgot_password", c.forgotPasswordAction)
}




type loginData struct {
	Domain string
	SupportNumber string
	IsUS bool
}

func (c *AuthController) loginAction(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {


		var data loginData
		data.Domain = os.Getenv("SAFE_STOP_DOMAIN")
		data.SupportNumber = os.Getenv("SAFE_STOP_SUPPORT_NUMBER")
		data.IsUS = (os.Getenv("SAFE_STOP_DOMAIN") == "safestopapp.com")


		c.render(w, r, "login", data)
	} else {
		r.ParseForm()
		// logic part of log in
		//fmt.Println("username:", r.Form["username"][0])
		//fmt.Println("password:", r.Form["password"][0])

		user := authenticateUser(string(r.Form["username"][0]), r.Form["password"][0])
		if user != nil {

			session, err := c.SessionStore.Get(r, "auth")
			if err != nil {
				//TODO: set flash message about bad cookie, tell user to clear cookies
				log.Println(err)
				http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
				return
			}

			//session.Values["userID"] = user.Id
			session.Values["current_user_email"] = user.Email

			err = session.Save(r, w)
			if err != nil {
				//TODO: set flash message about not saving session
				http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
				return
			}

			http.Redirect(w, r, r.URL.Host+"/dashboard", http.StatusFound)
			return
		}
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)

	}
}

func (c *AuthController) logoutAction(w http.ResponseWriter, r *http.Request) {

	session, err := c.SessionStore.Get(r, "auth")
	if err != nil {
		//TODO: set flash message about bad cookie, tell user to clear cookies
		log.Println(err)
		http.Redirect(w, r, r.URL.Host+"/", http.StatusFound)
		return
	}

	session.Values["current_user_email"] = nil

	err = session.Save(r, w)
	if err != nil {
		//TODO: set flash message about not saving session
		http.Redirect(w, r, r.URL.Host+"/", http.StatusFound)
		return
	}

	http.Redirect(w, r, r.URL.Host+"/", http.StatusFound)

}

type registrationFormData struct {
	Countries []models.Country
	Genders []string
	Orientations []string
}

func (c *AuthController) registerAction(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		var data registrationFormData
		data.Countries = models.AllCountries()
		c.render(w, r, "register", data)
	} else {
		r.ParseForm()

		//countryId, _ := strconv.Atoi(r.FormValue("country-id"))
		user := models.NewUser(r.FormValue("email"), r.FormValue("password"))
		if user != nil {

			session, err := c.SessionStore.Get(r, "auth")
			if err != nil {
				//TODO: set flash message about bad cookie, tell user to clear cookies
				log.Println(err)
				http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
				return
			}

			session.Values["current_user_email"] = user.Email
			err = session.Save(r, w)
			if err != nil {
				//TODO: set flash message about not saving session
				http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
				return
			}



			http.Redirect(w, r, r.URL.Host+"/dashboard", http.StatusFound)
		}
		http.Redirect(w, r, r.URL.Host+"/register", http.StatusFound)

	}
}

func (c *AuthController) forgotPasswordAction(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		c.render(w, r, "forgot_password", nil)
	} else {
		r.ParseForm()

		user := models.FindUserByEmail(string(r.Form["username"][0]))
		if user != nil {



			http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
			return
		}
		http.Redirect(w, r, r.URL.Host+"/forgot_password", http.StatusFound)

	}
}

func authenticateUser(email string, password string) *models.User {
	user := models.AuthenticateUser(email, password)
	if user != nil {
		return user
	}
	return nil
}
