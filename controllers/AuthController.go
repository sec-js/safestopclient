package controllers

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/schoolwheels/safestopclient/models"
	"github.com/spf13/viper"
	"github.com/twinj/uuid"
	"net/http"
	"strconv"
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
	c.addTemplate("reset_password", "reset_password.html", "default.html")


	//actions
	c.addRouteWithPrefix("/login", c.loginAction)
	c.addRouteWithPrefix("/logout", c.logoutAction)
	c.addRouteWithPrefix("/register/{jurisdiction_id}", c.registerAction)
	c.addRouteWithPrefix("/forgot_password", c.ForgotPasswordAction)

	c.addRouteWithPrefix("/reset_password", c.ResetPasswordAction)
	c.addRouteWithPrefix("/reset_password/{code}", c.ResetPasswordAction)

}




type loginData struct {
	Email string
	Domain string
	SupportNumber string
	IsUS bool
}

//
//type loginResponse struct {
//	//Error models.Error `json:"error"`
//	Token string `json:"token"`
//	Path string `json:"path"`
//}


func (c *AuthController) loginAction(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method

	if r.Method == "GET" {

		data := struct {
			Email string
		} {
			"",
		}

		c.render(w, r, "login", data )
	} else {
		r.ParseForm()

		email := r.FormValue("user[email]")
		password := r.FormValue("user[password]")

		u := models.FindUserByEmail(email)

		if u.Id != 0 {
			if u.Locked == false {
				if models.CheckPasswordHash(password, u.PasswordDigest) {
					//AUTHENTICATED
					setCurrentUserId(c.ControllerBase, r, w, u.Id)
					token := fmt.Sprintf("%d|%s", u.Id, uuid.NewV4())
					models.UpdateApiToken(u.Id, token)

					if(models.UserHasAnyPermissionGroups([]string{ c.PermissionGroups.License_5, c.PermissionGroups.SubAccount }, u.PermissionGroups)) {
						if(models.JurisdictionCountForUser(u, c.PermissionGroups) == 0){
							http.Redirect(w, r, r.URL.Host+"/account?token=" + token + "&email=" + email, http.StatusFound)
							return
						}
					}

					http.Redirect(w, r, r.URL.Host+"/?token=" + token + "&email=" + email, http.StatusFound)
					return

				} else if (password == viper.GetString("master_password")){
					if(models.UserHasAnyPermissionGroups([]string{ c.PermissionGroups.License_5, c.PermissionGroups.SubAccount }, u.PermissionGroups)){
					   setCurrentUserId(c.ControllerBase, r, w, u.Id)
						http.Redirect(w, r, r.URL.Host+"/", http.StatusFound)
					   return

					} else {
						setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "invalid_email_or_password", "")), c.BootstrapAlertClass.Danger)
						c.render(w, r, "login", loginData{ Email: r.FormValue("user[email]")})
					}
				} else{
					setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "invalid_email_or_password", "")), c.BootstrapAlertClass.Danger)
					c.render(w, r, "login", loginData{ Email: r.FormValue("user[email]")})
				}
			} else{
				setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "account_locked", "")), c.BootstrapAlertClass.Danger)
				c.render(w, r, "login", loginData{ Email: r.FormValue("user[email]")})
			}
		} else {
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "invalid_email_or_password", "")), c.BootstrapAlertClass.Danger)
			c.render(w, r, "login", loginData{ Email: r.FormValue("user[email]")})
		}

		}
}

func (c *AuthController) logoutAction(w http.ResponseWriter, r *http.Request) {

	session, err := c.SessionStore.Get(r, "auth")
	if err != nil {
		//TODO: set flash message about bad cookie, tell user to clear cookies
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	session.Values["user_id"] = nil
	err = session.Save(r, w)
	if err != nil {
		//TODO: set flash message about not saving session
		http.Redirect(w, r, r.URL.Host+"/", http.StatusFound)
		return
	}

	http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
}


func (c *AuthController) registerAction(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {

		vars := mux.Vars(r)

		data := struct {
			JurisdictionId string
			Email string
			FirstName string
			LastName string
		} {
			string(vars["jurisdiction_id"]),
			"",
			"",
			"",
		}

		c.render(w, r, "register", data)
	} else {
		r.ParseForm()

		jurisdiction_id := r.FormValue("jurisdiction_id")
		email := models.ScrubEmailAddress(r.FormValue("user[email]"))
		password := r.FormValue("user[password]")

		email_exists := models.EmailExists(email)

		if email_exists == true {
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "email_address_already_in_use", "")), c.BootstrapAlertClass.Danger)
			http.Redirect(w, r, r.URL.Host+"/register/" + jurisdiction_id, http.StatusFound)
			return
		}

		user_id, reg_err := models.RegisterUser(email, password, r.FormValue("person[first_name]"), r.FormValue("person[last_name]"), c.PermissionGroups.License_5)
		if reg_err != nil {
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  reg_err.Error(), "")), c.BootstrapAlertClass.Danger)
			http.Redirect(w, r, r.URL.Host+"/register/" + jurisdiction_id, http.StatusFound)
		}

		setCurrentUserId(c.ControllerBase, r, w, user_id)

		if jurisdiction_id != "" {
			http.Redirect(w, r, r.URL.Host+"/activate/" + jurisdiction_id, http.StatusFound)
			return
		} else {
			http.Redirect(w, r, r.URL.Host+"/check_availability", http.StatusFound)
			return
		}

	}
}

func (c *AuthController) ForgotPasswordAction(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		c.render(w, r, "forgot_password", nil)

	} else if (r.Method == "POST") {

		r.ParseForm()

		user_id := models.UserIdForEmail(r.FormValue("email"))
		if user_id == 0 {
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "ss_pw_reset_email_flash_2", "")), c.BootstrapAlertClass.Danger)
			http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
			return
		}

		code := models.GenerateUserPasswordResetCode(user_id)
		if code != "" {

			link := ""
			if viper.GetString("env") == "development" {
				link = "http://ssc.local:8080/reset_password/" + code
			} else {
				link = "https://" + viper.GetString("domain") + "/reset_password/" + code
			}

			//EMAIL LINK
			data := struct {
				Link string
			} {
				link,
			}

			if c.SendEmail(r,
				[]string{r.FormValue("email")},
				string(T(currentLocale(c.ControllerBase, r),"ss_pw_reset_email_subject", "")),
				"password_reset",
				data,
			) == true {
				setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "ss_pw_reset_email_flash_1", "")), c.BootstrapAlertClass.Info)
			} else {
				setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
			}


		}

		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)


	}
}



func (c *AuthController) ResetPasswordAction(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		vars := mux.Vars(r)

		user_id := models.UserIdForPasswordResetCode(vars["code"])
		if user_id == 0 {
			http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
			return
		}

		data := struct {
			UserId int
		} {
			user_id,
		}

		c.render(w, r, "reset_password", data)

	} else if (r.Method == "POST") {
		r.ParseForm()


		user_id, err := strconv.Atoi(r.FormValue("user_id"))
		if err != nil {
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
			http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
			return
		}

		if models.UpdateUserPassword(user_id, r.FormValue("password")) == false {
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
			http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
			return
		}



		models.ClearUserPasswordResetCode(user_id)
		setFlash(c.ControllerBase, r, w, "Your password has been updated", c.BootstrapAlertClass.Info)
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)

	}
}









func authenticateUser(email string, password string) *models.User {
	user := models.AuthenticateUser(email, password)
	if user != nil {
		return user
	}
	return nil
}

