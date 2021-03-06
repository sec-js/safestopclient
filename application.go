package main

import (
	"encoding/gob"
	"fmt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
	"github.com/schoolwheels/safestopclient/controllers"
	"github.com/schoolwheels/safestopclient/models"
	"github.com/spf13/viper"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)


var sessionStore = sessions.NewCookieStore([]byte("Byte my ass 2018!"))

func main() {

	gob.Register(controllers.FlashMessage{})

	var BootstrapAlertClass models.BootstrapAlertClass = models.BootstrapAlertClass{
		Primary: "primary",
		Secondary: "secondary",
		Success: "success",
		Danger: "danger",
		Warning: "warning",
		Info: "info",
		Light: "light",
		Dark: "dark",
	}

	var PermissionGroups models.PermissionGroups = models.PermissionGroups{
		Admin: "SafeStop Admin",
		License_1: "License 1 – Transportation Executive",
		License_2: "License 2 – Transportation Professional",
		License_3: "License 3 – SafeStop Administrator",
		License_4: "License 4 – SafeStop User Plus",
		License_5: "License 5 – SafeStop User",
		SubAccount: "SafeStop User Sub Account",
	}

	I18n := i18n.New(
		yaml.New(filepath.Join("./config/locales")), // load translations from the YAML files in directory `config/locales`
	)
	french := I18n.T("fr", "french")
	I18n.Default("Default").T("en", "french")
	fmt.Println("french: ",french)

	fmt.Println("~~~~~ SafeStop Client ~~~~~")
	//fmt.Println("SSC_ENV:", viper.GetString("env"))


	viper.SetEnvPrefix("SSC")
	viper.AutomaticEnv()
	viper.SetConfigName("config")       // name of config file (without extension)
	viper.AddConfigPath("./config") // look for config in the working directory
	err := viper.ReadInConfig()     // Find and read the config file
	if err != nil { // Handle errors reading the config file
		log.Fatal(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	fmt.Println("SSC_ENV:", viper.GetString("env"))


	fmt.Println(viper.GetString("db_host"))


	//sessions setup
	if viper.GetString("env") == "development" {
		sessionStore.Options = &sessions.Options{
			Domain: viper.GetString("domain"),
			Path:   "/",
			MaxAge: 3600 * 24,
			HttpOnly: true,
		}
	} else {
		sessionStore.Options = &sessions.Options{
			Domain: viper.GetString("domain"),
			Path:   "/",
			MaxAge: 3600 * 24,
			Secure:   true,
			HttpOnly: true,
		}
	}

	r := mux.NewRouter()

	//static files
	r.PathPrefix("/images/").Handler(http.FileServer(http.Dir("./public/")))
	r.PathPrefix("/media/").Handler(http.FileServer(http.Dir("./public/")))
	r.PathPrefix("/stylesheets/").Handler(http.FileServer(http.Dir("./public/")))
	r.PathPrefix("/javascript/").Handler(http.FileServer(http.Dir("./public/")))
	r.PathPrefix("/vendors/").Handler(http.FileServer(http.Dir("./public/")))


	//controllers
	//todo: possibly handle registration from package init methods in each controller's go file
	AuthController := controllers.AuthController{ &controllers.ControllerBase{Name: "AuthController", Templates: make(map[string]*template.Template), Router: r, SessionStore: sessionStore, BootstrapAlertClass: &BootstrapAlertClass, PermissionGroups: &PermissionGroups,}}
	AuthController.Register()

	AppController := controllers.AppController{&controllers.ControllerBase{Name: "AppController", Templates: make(map[string]*template.Template), Router: r, SessionStore: sessionStore, BootstrapAlertClass: &BootstrapAlertClass, PermissionGroups: &PermissionGroups,}}
	AppController.Register()

	APIController := controllers.APIController{&controllers.ControllerBase{Name: "APIController", Templates: make(map[string]*template.Template), Router: r, SessionStore: sessionStore, PermissionGroups: &PermissionGroups}}
	APIController.Register()

	AlertsController := controllers.AlertsController{&controllers.ControllerBase{Name: "AlertsController", Templates: make(map[string]*template.Template), Router: r, SessionStore: sessionStore, BootstrapAlertClass: &BootstrapAlertClass, PermissionGroups: &PermissionGroups,}}
	AlertsController.Register()

	http.Handle("/", r)
	if viper.GetString("env") == "development" {

		log.Fatal(http.ListenAndServe(":8080", csrf.Protect([]byte("32-byte-long-auth-key"), csrf.Secure(false))(r)))

	} else {
		// redirect every http request to https

		//go func() {
			if err := http.ListenAndServe(":5000", csrf.Protect([]byte("32-byte-long-auth-key"), csrf.Secure(true))(r) ); err != nil {
				log.Fatalf("ListenAndServe error: %v", err)
			}
		//}()

		//go http.ListenAndServe(":5000", middleware.RequestLogger(http.HandlerFunc(redirect)), )
		//log.Fatal(http.ListenAndServe(":443", csrf.Protect([]byte("32-byte-long-auth-key"), csrf.Secure(true))(r)))
		//log.Fatal(http.ListenAndServeTLS(":8443", "certs/safestopapp.com.pem", "certs/safestopapp.com-key.pem", middleware.RequestLogger(r)))
	}
}


//func redirect(w http.ResponseWriter, req *http.Request) {
//	target := req.TLS.
//	if req.URL.Scheme == "http" {
//		// remove/add not default ports from req.Host
//		target = "https://" + req.Host + req.URL.Path
//		if len(req.URL.RawQuery) > 0 {
//			target += "?" + req.URL.RawQuery
//		}
//		log.Printf("redirect to: %s", target)
//	}
//	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
//}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		//panic(err)
	}
}