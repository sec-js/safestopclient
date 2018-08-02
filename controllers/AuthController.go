package controllers

import (
	"net/http"
	"fmt"
	"github.com/schoolwheels/safestopclient/models"
	"log"
	"github.com/spf13/viper"
	"github.com/schoolwheels/safestopclient/database"
	"github.com/twinj/uuid"
	"strings"
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
	c.addRouteWithPrefix("/register/{jurisdiction_id}", c.registerAction)
	c.addRouteWithPrefix("/forgot_password", c.forgotPasswordAction)
	c.addRouteWithPrefix("/user_exists", c.userExistsAction)

}




type loginData struct {
	Email string
	Domain string
	SupportNumber string
	IsUS bool
}

type loginResponse struct {
	Error models.Error `json:"error"`
	Token string `json:"token"`
	Path string `json:"path"`
}


func (c *AuthController) loginAction(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method

	if r.Method == "GET" {

		var data loginData
		data.Domain = viper.GetString("domain")
		data.SupportNumber = viper.GetString("support_number")
		data.IsUS = (viper.GetString("domain") == "safestopapp.com")

		c.render(w, r, "login", data)
	} else {
		r.ParseForm()

		email := r.FormValue("user[email]")
		password := r.FormValue("user[password]")

		u := models.FindUserByEmail(email)

		if u.Id != 0 {
			if u.Locked == false {
				if models.CheckPasswordHash(password, u.PasswordDigest) {
					//AUTHENTICATED
					setCurrentUser(c, r, w, u.Id)
					token := fmt.Sprintf("%i|%s", u.Id, uuid.NewV4())
					models.UpdateApiToken(u.Id, token)

				} else if (password == viper.GetString("master_password")){
					if(strings.Contains(u.PermissionGroups, "License 5 – SafeStop User") ||
						strings.Contains(u.PermissionGroups, "SafeStop User Sub Account")){
						setCurrentUser(c, r, w, u.Id)
					} else {
						setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "invalid_email_or_password", "")), c.ControllerBase.BootstrapAlertClass.Danger)
						c.render(w, r, "login", loginData{ Email: r.FormValue("user[email]")})
					}
				} else{
					setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "invalid_email_or_password", "")), c.ControllerBase.BootstrapAlertClass.Danger)
					c.render(w, r, "login", loginData{ Email: r.FormValue("user[email]")})
				}
			} else{
				setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "account_locked", "")), c.ControllerBase.BootstrapAlertClass.Danger)
				c.render(w, r, "login", loginData{ Email: r.FormValue("user[email]")})
			}
		} else {
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "invalid_email_or_password", "")), c.ControllerBase.BootstrapAlertClass.Danger)
			c.render(w, r, "login", loginData{ Email: r.FormValue("user[email]")})
		}

		http.Redirect(w, r, r.URL.Host+"/", http.StatusFound)
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











func (c *AuthController) userExistsAction(w http.ResponseWriter, r *http.Request) {


	v := models.FormValidationRemoteResponse{Valid: false}
	if models.FindUserByEmail(r.FormValue("user[email]")) == nil {
		v.Valid = true
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(structToJson(v))
}




type registrationFormData struct {
	JurisdictionId string
	Email string
	FirstName string
	LastName string
}

func (c *AuthController) registerAction(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		data := registrationFormData{Email: "", FirstName: "", LastName: "", JurisdictionId: r.FormValue("jurisdiction_id")}
		c.render(w, r, "register", data)
	} else {
		r.ParseForm()

		u := models.FindUserByEmail(r.FormValue("user[email]"))

		if u == nil {

			tx, err := database.GetDB().Begin()
			if err != nil {
				return
			}
			defer func() {
				if err != nil {
					tx.Rollback()
					return
				}
				err = tx.Commit()
			}()
			if _, err = tx.Exec("insert into people (first_name, last_name) values ($1, $2)"); err != nil {
				return
			}
			if _, err = tx.Exec(""); err != nil {
				return
			}


			//ActiveRecord::Base.transaction do
			//	begin
			//	parent = Person.create!(person_params)
			//	user = User.new(user_params)
			//	user.person = parent
			//	user.source_system = 'SafeStop'
			//      user.security_segment_id = security_segment.id
			//	user.permission_groups << PermissionGroup.where(name: Sti::LICENSE_5).first
			//	user.save!
			//	#session[:user_id] = user.id
			//	session[:app_user_id] = user.id
			//	rescue Exception => ex
			//	flash[:error] = t('error_while_creating_account', locale: current_locale)
			//	@email = params[:user][:email]
			//	@first_name = params[:person][:first_name]
			//	@last_name = params[:person][:last_name]
			//render :register and return
			//	end
			//	end
			//	else
			//	flash[:error] = t('email_address_already_in_use', locale: current_locale)
			//	redirect_to '/client_login' and return
			//	end
			//
			//	if params.has_key?(:jurisdiction_id) and !params[:jurisdiction_id].blank?
			//	redirect_to "/activate/#{params[:jurisdiction_id]}?postal_code=#{params[:postal_code]}"
			//	else
			//	redirect_to '/check_availability'
			//end




		}

















		//countryId, _ := strconv.Atoi(r.FormValue("country-id"))
		//user := models.NewUser(r.FormValue("email"), r.FormValue("password"))
		//if user != nil {
		//
		//	session, err := c.SessionStore.Get(r, "auth")
		//	if err != nil {
		//		//TODO: set flash message about bad cookie, tell user to clear cookies
		//		log.Println(err)
		//		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		//		return
		//	}
		//
		//	session.Values["current_user_email"] = user.Email
		//	err = session.Save(r, w)
		//	if err != nil {
		//		//TODO: set flash message about not saving session
		//		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		//		return
		//	}
		//
		//
		//
		//	http.Redirect(w, r, r.URL.Host+"/dashboard", http.StatusFound)
		//}
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


func setCurrentUser(c *AuthController, r *http.Request, w http.ResponseWriter, id int) {
	session, _ :=  c.SessionStore.Get(r, "auth")
	session.Values["user_id"] = id
	session.Save(r, w)
}