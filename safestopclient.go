package main

import (
	"html/template"
	"net/http"
	"log"
	"fmt"
	"os"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	_ "github.com/lib/pq"
	"github.com/gorilla/sessions"
	"github.com/schoolwheels/safestopclient/controllers"
	"github.com/schoolwheels/safestopclient/middleware"

	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
	"path/filepath"
)

var sessionStore = sessions.NewCookieStore([]byte("Byte my ass 2018!"))

func main() {

	I18n := i18n.New(
		yaml.New(filepath.Join("./config/locales")), // load translations from the YAML files in directory `config/locales`
	)
	french := I18n.T("fr", "french")
	I18n.Default("Default").T("en", "french")
	fmt.Println("french: ",french)

	fmt.Println("~~~~~ SafeStop Client ~~~~~")
	fmt.Println("SSC_ENV:", os.Getenv("SSC_ENV"))

	viper.SetEnvPrefix("SSC")
	viper.AutomaticEnv()
	viper.SetConfigName("config")       // name of config file (without extension)
	viper.AddConfigPath("./config") // look for config in the working directory
	err := viper.ReadInConfig()     // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

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

	//controllers
	//todo: possibly handle registration from package init methods in each controller's go file
	AuthController := controllers.AuthController{&controllers.ControllerBase{Name: "AuthController", Templates: make(map[string]*template.Template), Router: r, SessionStore: sessionStore}}
	AuthController.Register()

	AppController := controllers.AppController{&controllers.ControllerBase{Name: "AppController", Templates: make(map[string]*template.Template), Router: r, SessionStore: sessionStore}}
	AppController.Register()

	http.Handle("/", r)
	log.Println("Listening...")
	if viper.GetString("env") == "development" {
		log.Fatal(http.ListenAndServe(":8080", middleware.RequestLogger(r)))
	} else {
		// redirect every http request to https
		go http.ListenAndServe(":8080", http.HandlerFunc(redirect))
		log.Fatal(http.ListenAndServeTLS(":8443", "certs/safestopapp.com.pem", "certs/safestopapp.com-key.pem", middleware.RequestLogger(r)))
	}
}


func redirect(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target,
		// see @andreiavrammsd comment: often 307 > 301
		http.StatusTemporaryRedirect)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}